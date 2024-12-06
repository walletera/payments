package tests

import (
    "context"
    "fmt"
    "log/slog"
    "testing"
    "time"

    "github.com/cucumber/godog"
    "github.com/walletera/message-processor/messages"
    "github.com/walletera/message-processor/rabbitmq"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/app"
    "github.com/walletera/payments/internal/domain/payment/event/handlers"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
)

const (
    postPaymentResponseKey = "postPaymentResponse"
    eventsMsgChKey         = "eventsMsgChKey"
    httpServerPort         = 8484
)

var testLogger *slog.Logger

func TestCreatePayment(t *testing.T) {

    zapLogger, err := newZapLogger()
    if err != nil {
        panic(err)
    }
    testLogger = slog.New(zapslog.NewHandler(zapLogger.Core(), nil))

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
    ctx.Step(`^a walletera customer$`, aWalleteraCustomer)
    ctx.Step(`^the customer sends the following payment to the payments endpoint:$`, theCustomerSendsTheFollowingPayment)
    ctx.Step(`^the endpoint returns the http status code 201$`, theEndpointReturnsTheHttpStatusCode201)
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

func aWalleteraCustomer(ctx context.Context) (context.Context, error) {
    return ctx, nil
}

func theCustomerSendsTheFollowingPayment(ctx context.Context, rawPayment *godog.DocString) (context.Context, error) {
    res, err := createPayment(ctx, rawPayment.Content)
    if err != nil {
        return ctx, err
    }
    return context.WithValue(ctx, postPaymentResponseKey, res), nil
}

func theEndpointReturnsTheHttpStatusCode201(ctx context.Context) (context.Context, error) {
    res, ok := ctx.Value(postPaymentResponseKey).(*api.Payment)
    if !ok {
        return ctx, fmt.Errorf("postPayment response is not what we expect: %v", res)
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

func newZapLogger() (*zap.Logger, error) {
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
    zapConfig := zap.Config{
        Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
        Development: false,
        Sampling: &zap.SamplingConfig{
            Initial:    100,
            Thereafter: 100,
        },
        Encoding:         "json",
        EncoderConfig:    encoderConfig,
        OutputPaths:      []string{"stderr"},
        ErrorOutputPaths: []string{"stderr"},
    }
    return zapConfig.Build()
}
