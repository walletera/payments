//go:build publicapi

package publicapi

import (
    "context"
    "fmt"
    "os"
    "testing"

    "github.com/cucumber/godog"
    "github.com/walletera/eventskit/messages"
    "github.com/walletera/eventskit/rabbitmq"
    api "github.com/walletera/payments-types/publicapi"
    "github.com/walletera/payments/internal/app"
    "github.com/walletera/payments/internal/domain/payment/event/handlers"
)

const (
    postPaymentResponseKey = "postPaymentResponse"
    eventsMsgChKey         = "eventsMsgChKey"
)

func TestCreatePayment(t *testing.T) {

    suite := godog.TestSuite{
        ScenarioInitializer: InitializeCreatePaymentScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features/create_payment.feature"},
            TestingT: t, // Testing instance that will run subtests.
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeCreatePaymentScenario(ctx *godog.ScenarioContext) {
    ctx.Before(beforeScenarioHook)
    ctx.Step(`^a running payments service$`, aRunningPaymentsService)
    ctx.Step(`^a running payments events consumer with queueName: "([^"]*)"$`, consumePaymentEvents)
    ctx.Step(`^an authorized walletera customer$`, anAuthorizedWalleteraCustomer)
    ctx.Step(`^an unauthorized walletera customer$`, anUnauthorizedWalleteraCustomer)
    ctx.Step(`^a walletera customer with an invalid token$`, aWalleteraCustomerWithAnInvalidToken)
    ctx.Step(`^the payment is created$`, thePaymentIsCreated)
    ctx.Step(`^the customer sends the following payment to the payments endpoint:$`, theCustomerSendsTheFollowingPayment)
    ctx.Step(`^the endpoint returns the http status code: (\d+)$`, theEndpointReturnsTheHttpStatusCode)
    ctx.Step(`^the payments service publish the following event:$`, thePaymentsServicePublishTheFollowingEvent)
    ctx.After(afterScenarioHook)
}

func consumePaymentEvents(ctx context.Context, queueName string) (context.Context, error) {
    client, err := rabbitmq.NewClient(
        rabbitmq.WithExchangeName(app.PaymentsServiceExchangeName),
        rabbitmq.WithExchangeType(app.PaymentServiceExchangeType),
        rabbitmq.WithQueueName(queueName),
        rabbitmq.WithConsumerRoutingKeys(handlers.PaymentCreatedRoutingKey, handlers.PaymentUpdatedRoutingKey),
    )
    if err != nil {
        return ctx, fmt.Errorf("failed creating rabbitmq client: %w", err)
    }
    msgCh, err := client.Consume()
    if err != nil {
        return ctx, fmt.Errorf("failed consuming messages from rabbitmq: %w", err)
    }
    ctx = context.WithValue(ctx, eventsMsgChKey, msgCh)
    return ctx, nil
}

func thePaymentIsCreated(ctx context.Context) (context.Context, error) {
    // TODO
    return ctx, nil
}

func theCustomerSendsTheFollowingPayment(ctx context.Context, paymentJsonFilePath *godog.DocString) (context.Context, error) {
    if paymentJsonFilePath == nil || len(paymentJsonFilePath.Content) == 0 {
        return ctx, fmt.Errorf("the paymentJsonFilePath is empty or was not defined")
    }

    paymentJson, err := os.ReadFile(paymentJsonFilePath.Content)
    if err != nil {
        return ctx, fmt.Errorf("error reading raw payment JSON file: %w", err)
    }

    res, err := createPayment(ctx, paymentJson, publicApiHttpServerPort)
    if err != nil {
        return ctx, err
    }
    return context.WithValue(ctx, postPaymentResponseKey, res), nil
}

func theEndpointReturnsTheHttpStatusCode(ctx context.Context, expectedStatusCode int) (context.Context, error) {
    postPaymentResponse := ctx.Value(postPaymentResponseKey)
    if postPaymentResponse == nil {
        return ctx, fmt.Errorf("postPayment response is not what we expect: %v", postPaymentResponse)
    }
    var responseStatusCode int
    switch postPaymentResponse.(type) {
    case *api.Payment:
        responseStatusCode = 201
    case *api.PostPaymentUnauthorized:
        responseStatusCode = 401
    case *api.PostPaymentBadRequest:
        responseStatusCode = 400
    case *api.PostPaymentConflict:
        responseStatusCode = 409
    default:
        return ctx, fmt.Errorf("postPayment response is not what we expect: %v", postPaymentResponse)
    }

    if responseStatusCode != expectedStatusCode {
        return ctx, fmt.Errorf("postPayment response is not what we expect: %v", postPaymentResponse)
    }
    return ctx, nil
}

func eventsMsgChFromCtx(ctx context.Context) <-chan messages.Message {
    value := ctx.Value(eventsMsgChKey)
    if value == nil {
        panic("missing eventsMsgCh on context")
    }
    ch, ok := value.(<-chan messages.Message)
    if !ok {
        panic("eventsMsgChKey has not the expected type")
    }
    return ch
}
