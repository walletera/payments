package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

const (
    expectationEndpoint = "/mockserver/expectation"
    verifyEndpoint      = "/mockserver/verify"
    clearEndpoint       = "/mockserver/clear"
)

type Client struct {
    baseURL    *url.URL
    httpClient IHTTPClient
}

func NewClient(baseURL *url.URL, httpClient IHTTPClient) *Client {
    return &Client{
        baseURL:    baseURL,
        httpClient: httpClient,
    }
}

// CreateExpectation creates an expectation
func (c *Client) CreateExpectation(ctx context.Context, expectation []byte) error {
    req, err := http.NewRequestWithContext(
        ctx,
        http.MethodPut,
        c.baseURL.JoinPath(expectationEndpoint).String(),
        bytes.NewReader(expectation),
    )
    if err != nil {
        return fmt.Errorf("faile creating request for endpoint %s: %w", expectationEndpoint, err)
    }
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("expectation request failed: %w", err)
    }
    defer resp.Body.Close()
    _, err = io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed reading expectation request response body: %w", err)
    }
    if resp.StatusCode != http.StatusCreated {
        switch resp.StatusCode {
        case http.StatusBadRequest:
            return &IncorrectRequestFormat{}
        case http.StatusNotAcceptable:
            return &InvalidExpectation{}
        default:
            return &UnexpectedStatusCode{
                Endpoint:   expectationEndpoint,
                StatusCode: resp.StatusCode,
            }
        }
    }
    return nil
}

// VerifyRequest verify a request has been received a specific number of times
func (c *Client) VerifyRequest(ctx context.Context, body VerifyRequestBody) error {
    marshalledBody, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("failed marshalling verify request body: %w", err)
    }
    req, err := http.NewRequestWithContext(
        ctx,
        http.MethodPut,
        c.baseURL.JoinPath(verifyEndpoint).String(),
        bytes.NewReader(marshalledBody),
    )
    if err != nil {
        return fmt.Errorf("faile creating request for endpoint %s: %w", verifyEndpoint, err)
    }
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("verify request failed: %w", err)
    }
    defer resp.Body.Close()
    _, err = io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed reading verify request response body: %w", err)
    }
    if resp.StatusCode != http.StatusAccepted {
        switch resp.StatusCode {
        case http.StatusBadRequest:
            return &IncorrectRequestFormat{}
        case http.StatusNotAcceptable:
            return &RequestHasNotBeenReceived{}
        default:
            return &UnexpectedStatusCode{
                Endpoint:   expectationEndpoint,
                StatusCode: resp.StatusCode,
            }
        }
    }
    return nil
}

// Clear clears all expectations and recorded requests that match the request matcher
func (c *Client) Clear(ctx context.Context) error {
    req, err := http.NewRequestWithContext(
        ctx,
        http.MethodPut,
        c.baseURL.JoinPath(clearEndpoint).String(),
        nil,
    )
    if err != nil {
        return fmt.Errorf("failed creating request for endpoint %s: %w", clearEndpoint, err)
    }
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("clear request failed: %w", err)
    }
    defer resp.Body.Close()
    _, err = io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed reading clear request response body: %w", err)
    }
    if resp.StatusCode != http.StatusOK {
        switch resp.StatusCode {
        case http.StatusBadRequest:
            return &IncorrectRequestFormat{}
        default:
            return &UnexpectedStatusCode{
                Endpoint:   expectationEndpoint,
                StatusCode: resp.StatusCode,
            }
        }
    }
    return nil
}
