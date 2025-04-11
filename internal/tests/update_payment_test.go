package tests

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/cucumber/godog"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/tests/httpauth"
    "github.com/walletera/payments/pkg/wuuid"
)

func TestUpdatePayment(t *testing.T) {

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
    ctx.Step(`^a running payments service$`, aRunningPaymentsService)
    ctx.Step(`^a running payments events consumer with queueName: "([^"]*)"$`, consumePaymentEvents)
    ctx.Step(`^an authorized walletera internal service$`, anAuthorizedWalleteraInternalService)
    ctx.Step(`^a payment in pending status$`, aPaymentInPendingStatus)
    ctx.Step(`^the payments service receive a PATCH request to update the payment to status: "([^"]*)"$`, thePaymentsServiceReceiveAPATCHRequestToUpdateThePayment)
    ctx.Step(`^the payment is updated to status: "([^"]*)"$`, thePaymentIsUpdatedToStatus)
    ctx.Step(`^the payments service publish the following event:$`, thePaymentsServicePublishTheFollowingEvent)
    ctx.After(afterScenarioHook)
}

func anAuthorizedWalleteraInternalService(ctx context.Context) (context.Context, error) {
    // TODO internal service JWT token
    return context.WithValue(ctx, authTokenKey, "some.valid.jwt"), nil
}

func aPaymentInPendingStatus(ctx context.Context) (context.Context, error) {
    paymentStr := `
    {
      "id": "%s",
      "customerId": "%s",
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
    paymentJson := fmt.Sprintf(paymentStr, wuuid.NewUUID(), wuuid.NewUUID())
    resp, err := createPayment(ctx, paymentJson, privateApiHttpServerPort)
    if err != nil {
        return ctx, fmt.Errorf("error creating payment: %v", err)
    }

    payment, ok := resp.(*api.Payment)
    if !ok {
        return ctx, fmt.Errorf("postPayment response's type is not *api.Payment: %T", resp)
    }

    return context.WithValue(ctx, paymentCreatedKey, payment), nil
}

func thePaymentIsUpdatedToStatus(ctx context.Context, status string) (context.Context, error) {
    payment := ctx.Value(paymentCreatedKey).(*api.Payment)
    paymentsClient, err := api.NewClient(
        fmt.Sprintf("http://127.0.0.1:%d", privateApiHttpServerPort),
        httpauth.NewSecuritySource(authTokenFromCtx(ctx)),
    )
    if err != nil {
        return nil, err
    }
    requestCtx, _ := context.WithTimeout(ctx, 200*time.Second)
    resp, err := paymentsClient.GetPayment(requestCtx, api.GetPaymentParams{
        PaymentId: payment.ID,
    })
    if err != nil {
        return nil, err
    }

    retrievedPayment, ok := resp.(*api.Payment)
    if !ok {
        return ctx, fmt.Errorf("Payment response is not of type *api.Payment: %v", resp)
    }

    if !retrievedPayment.Status.Set {
        return ctx, fmt.Errorf("retrieved payment status is not set")
    }

    if string(retrievedPayment.Status.Value) != status {
        return ctx, fmt.Errorf("retrieved payment status is not %s, instead it is %s", status, retrievedPayment.Status.Value)
    }

    return ctx, nil
}
