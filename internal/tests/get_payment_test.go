package tests

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "testing"
    "time"

    "github.com/cucumber/godog"
    "github.com/google/uuid"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/tests/httpauth"
    "go.uber.org/zap/exp/zapslog"
)

const getPaymentResponse = "getPaymentResponseFromCtx"

func TestGetPayment(t *testing.T) {

    zapLogger, err := newZapLogger()
    if err != nil {
        panic(err)
    }
    testLogger = slog.New(zapslog.NewHandler(zapLogger.Core(), nil))

    suite := godog.TestSuite{
        ScenarioInitializer: InitializeGetPaymentScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features/get_payment.feature"},
            TestingT: t, // Testing instance that will run subtests.
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeGetPaymentScenario(ctx *godog.ScenarioContext) {
    ctx.Before(beforeScenarioHook)
    ctx.Step(`^a running payments service$`, aRunningPaymentsService)
    ctx.Step(`^an authorized walletera customer$`, anAuthorizedWalleteraCustomer)
    ctx.Step(`^the following payment:$`, theFollowingPayment)
    ctx.Step(`^the payments service receive a PATCH request to update the payment to status: "([^"]*)"$`, thePaymentsServiceReceiveAPATCHRequestToUpdateThePayment)
    ctx.Step(`^the payments service receive a GET request to retrieve the payment with id: "([^"]*)"$`, thePaymentsServiceReceivesAGETRequestToRetrieveThePayment)
    ctx.Step(`^the payments service returns the following response:$`, thePaymentsServiceReturnsTheFollowingResponse)
    ctx.Step(`^the payments service returns the following status code:$`, thePaymentsServiceReturnsTheFollowingResponse)
    ctx.Step(`^the payments service returns 404 Not Found$`, thePaymentsServiceReturnsNotFoundStatusCode)
    ctx.After(afterScenarioHook)
}

func theFollowingPayment(ctx context.Context, paymentJson *godog.DocString) (context.Context, error) {
    resp, err := createPayment(ctx, paymentJson.Content, publicApiHttpServerPort)
    if err != nil {
        return ctx, err
    }

    payment, ok := resp.(*api.Payment)
    if !ok {
        return ctx, fmt.Errorf("postPayment response is not what we expect: %v", resp)
    }

    return context.WithValue(ctx, paymentCreatedKey, payment), nil
}

func thePaymentsServiceReceivesAGETRequestToRetrieveThePayment(ctx context.Context, paymentIdStr string) (context.Context, error) {
    paymentsClient, err := api.NewClient(
        fmt.Sprintf("http://127.0.0.1:%d", publicApiHttpServerPort),
        httpauth.NewSecuritySource(authTokenFromCtx(ctx)),
    )
    if err != nil {
        return nil, err
    }
    requestCtx, _ := context.WithTimeout(ctx, 5*time.Second)
    resp, err := paymentsClient.GetPayment(requestCtx, api.GetPaymentParams{PaymentId: uuid.MustParse(paymentIdStr)})
    if err != nil {
        return nil, err
    }
    return context.WithValue(ctx, getPaymentResponse, resp), nil
}

func thePaymentsServiceReturnsTheFollowingResponse(ctx context.Context, responseMatcher *godog.DocString) (context.Context, error) {
    ctx, err := createJSONMatcher(ctx, "thePaymentsServiceReturnsTheFollowingResponse", responseMatcher.Content)
    if err != nil {
        return ctx, err
    }
    resp := getPaymentResponseFromCtx(ctx)
    payment, ok := resp.(*api.Payment)
    if !ok {
        return ctx, fmt.Errorf("GET payment response should be *api.Payment but is: %+v", resp)
    }
    paymentJson, err := json.Marshal(payment)
    if err != nil {
        return ctx, err
    }
    matched, err := matchJSON(ctx, "thePaymentsServiceReturnsTheFollowingResponse", paymentJson)
    if err != nil {
        return ctx, err
    }
    if !matched {
        return ctx, fmt.Errorf("payment response didn't match expected response")
    }
    return ctx, nil
}

func thePaymentsServiceReturnsNotFoundStatusCode(ctx context.Context) (context.Context, error) {
    resp := getPaymentResponseFromCtx(ctx)
    _, ok := resp.(*api.GetPaymentNotFound)
    if !ok {
        return ctx, fmt.Errorf("GET payment response should be *api.GetPaymentNotFound but is: %+v", resp)
    }
    return ctx, nil
}

func getPaymentResponseFromCtx(ctx context.Context) api.GetPaymentRes {
    return ctx.Value(getPaymentResponse).(api.GetPaymentRes)
}
