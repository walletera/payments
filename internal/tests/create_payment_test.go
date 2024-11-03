package tests

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "testing"
    "time"

    "github.com/cucumber/godog"
    "github.com/walletera/message-processor/messages"
    "github.com/walletera/message-processor/rabbitmq"
    msClient "github.com/walletera/mockserver-go-client/pkg/client"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/app"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
    "golang.org/x/sync/errgroup"
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
        ScenarioInitializer: InitializeProcessWithdrawalCreatedScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features"},
            TestingT: t, // Testing instance that will run subtests.
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeProcessWithdrawalCreatedScenario(ctx *godog.ScenarioContext) {
    ctx.Before(beforeScenarioHook)
    ctx.Before(consumePaymentEvents)
    ctx.Step(`^a running payments service$`, aRunningPaymentsService)
    ctx.Step(`^a walletera customer$`, aWalleteraCustomer)
    ctx.Step(`^the customer sends the following payment to the payments endpoint:$`, theCustomerSendsTheFollowingPayment)
    ctx.Step(`^the endpoint returns the http status code 201$`, theEndpointReturnsTheHttpStatusCode201)
    ctx.Step(`^the payments service publish the following event:$`, thePaymentsServicePublishTheFollowingEvent)
    ctx.After(afterScenarioHook)
}

func consumePaymentEvents(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
    client, err := rabbitmq.NewClient(
        rabbitmq.WithExchangeName(app.PaymentsServiceExchangeName),
        rabbitmq.WithExchangeType(app.PaymentServiceExchangeType),
        rabbitmq.WithQueueName("createPaymentTestQueue"),
        rabbitmq.WithConsumerRoutingKeys(app.PaymentCreatedRoutingKey),
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
    paymentsClient, err := api.NewClient(fmt.Sprintf("http://127.0.0.1:%d", httpServerPort))
    if err != nil {
        return ctx, err
    }
    var payment api.Payment
    err = json.Unmarshal([]byte(rawPayment.Content), &payment)
    if err != nil {
        return ctx, fmt.Errorf("failed unmarshalling expected payment: %w", err)
    }
    requestCtx, _ := context.WithTimeout(ctx, 200*time.Second)
    res, err := paymentsClient.PostPayment(requestCtx, &payment, api.PostPaymentParams{})
    if err != nil {
        return ctx, err
    }
    return context.WithValue(ctx, postPaymentResponseKey, res), nil
}

func theEndpointReturnsTheHttpStatusCode201(ctx context.Context) (context.Context, error) {
    res, ok := ctx.Value(postPaymentResponseKey).(*api.PostPaymentCreated)
    if !ok {
        return ctx, fmt.Errorf("postPayment response is what we expect: %v", res)
    }
    return ctx, nil
}

func thePaymentsServicePublishTheFollowingEvent(ctx context.Context, eventMatcher *godog.DocString) (context.Context, error) {
    ch := eventsMsgChFromCtx(ctx)
    timeout := time.After(200 * time.Second)
    select {
    case <-timeout:
        return ctx, fmt.Errorf("timeout waiting for event to be published")
    case msg := <-ch:
        msg.Acknowledger().Ack()
        testLogger.Debug("[TESTLOG] published message", slog.String("message", string(msg.Payload())))
        matched, err := matchEvent(ctx, msg.Payload(), eventMatcher.Content)
        if err != nil {
            return ctx, fmt.Errorf("error matching published event: %w", err)
        }
        if !matched {
            return ctx, fmt.Errorf("published event %s don't match the expected event", msg.Payload())
        }
    }
    return ctx, nil
}

func matchEvent(ctx context.Context, payload []byte, matcher string) (bool, error) {
    ctx, err := createEventMatcher(ctx, "matchPaymentCreatedEvent", matcher)
    if err != nil {
        return false, err
    }
    _, err = http.Post(mockserverUrl+"/matchevent", "application/json", bytes.NewReader(payload))
    if err != nil {
        return false, err
    }
    err = verifyExpectationMetWithin(ctx, "matchPaymentCreatedEvent", 2*time.Second)
    if err != nil {
        return false, err
    }
    return true, nil
}

func verifyExpectationMetWithin(ctx context.Context, expectationID string, timeout time.Duration) error {
    errGroup := new(errgroup.Group)
    timeoutCh := time.After(timeout)
    errGroup.Go(func() error {
        var err error
        for {
            select {
            case <-timeoutCh:
                return fmt.Errorf("expectation %s was not met whithin %s: %w", expectationID, timeout.String(), err)
            default:
                err = verifyExpectationMet(ctx, expectationID)
                if err == nil {
                    return nil
                }
                time.Sleep(1 * time.Second)
            }
        }
    })
    return errGroup.Wait()
}

func verifyExpectationMet(ctx context.Context, expectationID string) error {
    verificationErr := mockServerClient().VerifyRequest(ctx, msClient.VerifyRequestBody{
        ExpectationId: msClient.ExpectationId{
            Id: expectationID,
        },
    })
    if verificationErr != nil {
        return verificationErr
    }
    return nil
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
