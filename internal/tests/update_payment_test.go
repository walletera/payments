package tests

import (
    "context"
    "fmt"
    "log/slog"
    "testing"
    "time"

    "github.com/cucumber/godog"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/pkg/wuuid"
    "go.uber.org/zap/exp/zapslog"
)

func TestUpdatePayment(t *testing.T) {

    zapLogger, err := newZapLogger()
    if err != nil {
        panic(err)
    }
    testLogger = slog.New(zapslog.NewHandler(zapLogger.Core(), nil))

    suite := godog.TestSuite{
        ScenarioInitializer: InitializeUpdatePaymentScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features/update_payment.feature"},
            TestingT: t, // Testing instance that will run subtests.
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeUpdatePaymentScenario(ctx *godog.ScenarioContext) {
    ctx.Before(beforeScenarioHook)
    ctx.Before(consumePaymentEvents)
    ctx.Step(`^a running payments service$`, aRunningPaymentsService)
    ctx.Step(`^a payment in pending status$`, aPaymentInPendingStatus)
    ctx.Step(`^the payments service receive a PATCH request to update the payment to status: "([^"]*)"$`, thePaymentsServiceReceiveAPATCHRequestToUpdateThePayment)
    ctx.Step(`^the payment is updated to status: "([^"]*)"$`, thePaymentIsUpdatedToStatus)
    ctx.Step(`^the payments service publish the following event:$`, thePaymentsServicePublishTheFollowingEvent)
    ctx.After(afterScenarioHook)
}

var paymentKey = "paymentKey"

func aPaymentInPendingStatus(ctx context.Context) (context.Context, error) {
    paymentStr := `
    {
      "amount": 100,
      "currency": "ARS",
      "beneficiary": {
        "bankName": "dinopay",
        "bankId": "dinopay",
        "accountHolder": "John Doe",
        "routingKey": "123456789123456"
      }
    }
`
    resp, err := createPayment(ctx, paymentStr)
    if err != nil {
        return ctx, err
    }

    payment, ok := resp.(*api.Payment)
    if !ok {
        return ctx, fmt.Errorf("postPayment response is not what we expect: %v", resp)
    }

    return context.WithValue(ctx, paymentKey, payment), nil
}

func thePaymentsServiceReceiveAPATCHRequestToUpdateThePayment(ctx context.Context, status string) (context.Context, error) {
    payment := ctx.Value(paymentKey).(*api.Payment)
    paymentsClient, err := api.NewClient(fmt.Sprintf("http://127.0.0.1:%d", httpServerPort))
    if err != nil {
        return nil, err
    }
    requestCtx, _ := context.WithTimeout(ctx, 200*time.Second)
    _, err = paymentsClient.PatchPayment(requestCtx, &api.PaymentUpdate{
        PaymentId: payment.ID.Value,
        ExternalId: api.OptUUID{
            Value: wuuid.NewUUID(),
            Set:   true,
        },
        Status: api.PaymentStatus(status),
    }, api.PatchPaymentParams{
        XWalleteraCorrelationID: api.OptUUID{
            Value: wuuid.NewUUID(),
            Set:   true,
        },
        PaymentId: payment.ID.Value,
    })
    if err != nil {
        return nil, err
    }
    return ctx, err
}

func thePaymentIsUpdatedToStatus(ctx context.Context, status string) (context.Context, error) {
    payment := ctx.Value(paymentKey).(*api.Payment)
    paymentsClient, err := api.NewClient(fmt.Sprintf("http://127.0.0.1:%d", httpServerPort))
    if err != nil {
        return nil, err
    }
    requestCtx, _ := context.WithTimeout(ctx, 200*time.Second)
    resp, err := paymentsClient.GetPayment(requestCtx, api.GetPaymentParams{
        PaymentId: payment.ID.Value,
    })
    if err != nil {
        return nil, err
    }

    retrievedPayment, ok := resp.(*api.Payment)
    if !ok {
        return ctx, fmt.Errorf("GetPayment response is not of type *api.Payment: %v", resp)
    }

    if !retrievedPayment.Status.Set {
        return ctx, fmt.Errorf("retrieved payment status is not set")
    }

    if string(retrievedPayment.Status.Value) != status {
        return ctx, fmt.Errorf("retrieved payment status is not %s, instead it is %s", status, retrievedPayment.Status.Value)
    }

    return ctx, nil
}
