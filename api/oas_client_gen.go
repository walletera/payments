// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/ogen-go/ogen/conv"
	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/otelogen"
	"github.com/ogen-go/ogen/uri"
)

// Invoker invokes operations described by OpenAPI v3 specification.
type Invoker interface {
	// PatchWithdrawal invokes patchWithdrawal operation.
	//
	// Patches a withdrawal.
	//
	// PATCH /withdrawals/{withdrawalId}
	PatchWithdrawal(ctx context.Context, request *WithdrawalPatchBody, params PatchWithdrawalParams) (PatchWithdrawalRes, error)
	// PostDeposit invokes postDeposit operation.
	//
	// Creates a deposit.
	//
	// POST /deposits
	PostDeposit(ctx context.Context, request *DepositPostBody) (PostDepositRes, error)
}

// Client implements OAS client.
type Client struct {
	serverURL *url.URL
	baseClient
}

var _ Handler = struct {
	*Client
}{}

func trimTrailingSlashes(u *url.URL) {
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = strings.TrimRight(u.RawPath, "/")
}

// NewClient initializes new Client defined by OAS.
func NewClient(serverURL string, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	trimTrailingSlashes(u)

	c, err := newClientConfig(opts...).baseClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		serverURL:  u,
		baseClient: c,
	}, nil
}

type serverURLKey struct{}

// WithServerURL sets context key to override server URL.
func WithServerURL(ctx context.Context, u *url.URL) context.Context {
	return context.WithValue(ctx, serverURLKey{}, u)
}

func (c *Client) requestURL(ctx context.Context) *url.URL {
	u, ok := ctx.Value(serverURLKey{}).(*url.URL)
	if !ok {
		return c.serverURL
	}
	return u
}

// PatchWithdrawal invokes patchWithdrawal operation.
//
// Patches a withdrawal.
//
// PATCH /withdrawals/{withdrawalId}
func (c *Client) PatchWithdrawal(ctx context.Context, request *WithdrawalPatchBody, params PatchWithdrawalParams) (PatchWithdrawalRes, error) {
	res, err := c.sendPatchWithdrawal(ctx, request, params)
	return res, err
}

func (c *Client) sendPatchWithdrawal(ctx context.Context, request *WithdrawalPatchBody, params PatchWithdrawalParams) (res PatchWithdrawalRes, err error) {
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("patchWithdrawal"),
		semconv.HTTPMethodKey.String("PATCH"),
		semconv.HTTPRouteKey.String("/withdrawals/{withdrawalId}"),
	}

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(float64(elapsedDuration)/float64(time.Millisecond)), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, "PatchWithdrawal",
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [2]string
	pathParts[0] = "/withdrawals/"
	{
		// Encode "withdrawalId" parameter.
		e := uri.NewPathEncoder(uri.PathEncoderConfig{
			Param:   "withdrawalId",
			Style:   uri.PathStyleSimple,
			Explode: false,
		})
		if err := func() error {
			return e.EncodeValue(conv.UUIDToString(params.WithdrawalId))
		}(); err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		encoded, err := e.Result()
		if err != nil {
			return res, errors.Wrap(err, "encode path")
		}
		pathParts[1] = encoded
	}
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "PATCH", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodePatchWithdrawalRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodePatchWithdrawalResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}

// PostDeposit invokes postDeposit operation.
//
// Creates a deposit.
//
// POST /deposits
func (c *Client) PostDeposit(ctx context.Context, request *DepositPostBody) (PostDepositRes, error) {
	res, err := c.sendPostDeposit(ctx, request)
	return res, err
}

func (c *Client) sendPostDeposit(ctx context.Context, request *DepositPostBody) (res PostDepositRes, err error) {
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("postDeposit"),
		semconv.HTTPMethodKey.String("POST"),
		semconv.HTTPRouteKey.String("/deposits"),
	}

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedDuration := time.Since(startTime)
		c.duration.Record(ctx, float64(float64(elapsedDuration)/float64(time.Millisecond)), metric.WithAttributes(otelAttrs...))
	}()

	// Increment request counter.
	c.requests.Add(ctx, 1, metric.WithAttributes(otelAttrs...))

	// Start a span for this request.
	ctx, span := c.cfg.Tracer.Start(ctx, "PostDeposit",
		trace.WithAttributes(otelAttrs...),
		clientSpanKind,
	)
	// Track stage for error reporting.
	var stage string
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			c.errors.Add(ctx, 1, metric.WithAttributes(otelAttrs...))
		}
		span.End()
	}()

	stage = "BuildURL"
	u := uri.Clone(c.requestURL(ctx))
	var pathParts [1]string
	pathParts[0] = "/deposits"
	uri.AddPathParts(u, pathParts[:]...)

	stage = "EncodeRequest"
	r, err := ht.NewRequest(ctx, "POST", u)
	if err != nil {
		return res, errors.Wrap(err, "create request")
	}
	if err := encodePostDepositRequest(request, r); err != nil {
		return res, errors.Wrap(err, "encode request")
	}

	stage = "SendRequest"
	resp, err := c.cfg.Client.Do(r)
	if err != nil {
		return res, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	stage = "DecodeResponse"
	result, err := decodePostDepositResponse(resp)
	if err != nil {
		return res, errors.Wrap(err, "decode response")
	}

	return result, nil
}
