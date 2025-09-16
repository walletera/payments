//go:build privateapi

package privateapi

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/cucumber/godog"
    privapi "github.com/walletera/payments-types/privateapi"
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
    ctx.Step(`^an authorized dinopay-gateway service$`, anAuthorizedWalleteraInternalService)
    ctx.Step(`^a payment is created in pending status$`, aPaymentIsCreatedInPendingStatus)
    ctx.Step(`^the payments service receive a PATCH request to update the payment to status: "([^"]*)"$`, thePaymentsServiceReceiveAPATCHRequestToUpdateThePayment)
    ctx.Step(`^the payment is updated to status: "([^"]*)"$`, thePaymentIsUpdatedToStatus)
    ctx.Step(`^the payments service publish the following event:$`, thePaymentsServicePublishTheFollowingEvent)
    ctx.After(afterScenarioHook)
}

func aPaymentIsCreatedInPendingStatus(ctx context.Context) (context.Context, error) {
    paymentStr := `
    {
      "id": "%s",
      "customerId": "%s",
      "amount": 100,
      "currency": "ARS",
      "gateway": "bind",
      "direction": "inbound",
      "status": "pending",
      "debtor": {
        "institutionName": "dinopay",
        "institutionId": "dinopay",
        "currency": "ARS",
        "accountDetails": {
          "accountType": "cvu",
          "cuit": "23679876453",
          "routingInfo": {
            "cvuRoutingInfoType": "cvu",
            "cvu": "1122334455667788554433"
          }
        }
      },
      "beneficiary": {
        "institutionName": "dinopay",
        "institutionId": "dinopay",
        "currency": "ARS",
        "accountDetails": {
          "accountType": "cvu",
          "cuit": "23679876453",
           "routingInfo": {
            "cvuRoutingInfoType": "cvu",
            "cvu": "1122334455667788554433"
          }
        }
      }
    }
`
    paymentJson := fmt.Sprintf(paymentStr, wuuid.NewUUID(), wuuid.NewUUID())
    resp, err := createPayment(ctx, paymentJson, privateApiHttpServerPort)
    if err != nil {
        return ctx, fmt.Errorf("error creating payment: %v", err)
    }

    payment, ok := resp.(*privapi.Payment)
    if !ok {
        return ctx, fmt.Errorf("postPayment response's type is not *api.Payment: %T", resp)
    }

    return context.WithValue(ctx, paymentCreatedKey, payment), nil
}

func thePaymentIsUpdatedToStatus(ctx context.Context, status string) (context.Context, error) {
    payment := ctx.Value(paymentCreatedKey).(*privapi.Payment)
    paymentsClient, err := privapi.NewClient(
        fmt.Sprintf("http://127.0.0.1:%d", privateApiHttpServerPort),
    )
    if err != nil {
        return nil, err
    }
    requestCtx, ctxCancel := context.WithTimeout(ctx, 200*time.Second)
    defer ctxCancel()
    resp, err := paymentsClient.GetPayment(requestCtx, privapi.GetPaymentParams{
        PaymentId: payment.ID,
    })
    if err != nil {
        return nil, err
    }

    retrievedPayment, ok := resp.(*privapi.Payment)
    if !ok {
        return ctx, fmt.Errorf("payment response is not of type *api.Payment: %v", resp)
    }

    if string(retrievedPayment.Status) != status {
        return ctx, fmt.Errorf("retrieved payment status is not %s, instead it is %s", status, retrievedPayment.Status)
    }

    return ctx, nil
}
