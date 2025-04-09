package tests

import (
    "bytes"
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "log/slog"
    "math/rand"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "time"

    "github.com/cucumber/godog"
    "github.com/golang-jwt/jwt"
    "github.com/walletera/eventskit/rabbitmq"
    slogwatcher "github.com/walletera/logs-watcher/slog"
    msClient "github.com/walletera/mockserver-go-client/pkg/client"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/app"
    "github.com/walletera/payments/internal/tests/httpauth"
    "github.com/walletera/payments/pkg/auth"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
    "golang.org/x/sync/errgroup"
)

const (
    eventStoreDBUrl           = "esdb://localhost:2113?tls=false"
    mockserverUrl             = "http://localhost:2090"
    publicApiHttpServerPort   = 8484
    privateApiHttpServerPort  = 8585
    paymentCreatedKey         = "paymentCreatedKey"
    appCtxCancelFuncKey       = "appCtxCancelFuncKey"
    logsWatcherKey            = "logsWatcher"
    authTokenKey              = "authToken"
    customerIdKey             = "customerIdKey"
    logsWatcherWaitForTimeout = 5 * time.Second
)

type MockServerExpectation struct {
    ExpectationID string `json:"id"`
}

func beforeScenarioHook(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
    handler, err := newZapHandler()
    if err != nil {
        return ctx, err
    }
    logsWatcher := slogwatcher.NewWatcher(handler)
    ctx = context.WithValue(ctx, logsWatcherKey, logsWatcher)
    return ctx, nil
}

func afterScenarioHook(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {

    logsWatcher := logsWatcherFromCtx(ctx)

    appCtxCancelFuncFromCtx(ctx)()
    foundLogEntry := logsWatcher.WaitFor("payments service stopped", logsWatcherWaitForTimeout)
    if !foundLogEntry {
        return ctx, fmt.Errorf("app termination failed (didn't find expected log entry)")
    }

    err = logsWatcher.Stop()
    if err != nil {
        return ctx, fmt.Errorf("failed stopping the logsWatcher: %w", err)
    }

    return ctx, nil
}

func aRunningPaymentsService(ctx context.Context) (context.Context, error) {

    logHandler := logsWatcherFromCtx(ctx).DecoratedHandler()
    appCtx, appCtxCancelFunc := context.WithCancel(ctx)

    base64Key, err := os.ReadFile("testdata/rsa_pub_key_base64")
    if err != nil {
        return ctx, err
    }

    go func() {
        app, err := app.NewApp(
            app.WithBase64AuthPubKey(string(base64Key)),
            app.WithRabbitmqHost(rabbitmq.DefaultHost),
            app.WithRabbitmqPort(rabbitmq.DefaultPort),
            app.WithRabbitmqUser(rabbitmq.DefaultUser),
            app.WithRabbitmqPassword(rabbitmq.DefaultPassword),
            app.WithESDBUrl(eventStoreDBUrl),
            app.WithPublicAPIHttpServerPort(publicApiHttpServerPort),
            app.WithPrivateAPIHttpServerPort(privateApiHttpServerPort),
            app.WithLogHandler(logHandler),
        )
        if err != nil {
            panic("failed initializing app: " + err.Error())
        }
        err = app.Run(appCtx)
        if err != nil {
            panic("failed running app" + err.Error())
        }
    }()

    ctx = context.WithValue(ctx, appCtxCancelFuncKey, appCtxCancelFunc)

    foundLogEntry := logsWatcherFromCtx(ctx).WaitFor("payments service started", logsWatcherWaitForTimeout)
    if !foundLogEntry {
        return ctx, fmt.Errorf("app startup failed (didn't find expected log entry)")
    }

    return ctx, nil
}

func aWalleteraCustomerWithAnInvalidToken(ctx context.Context) (context.Context, error) {
    customerId := wuuid.NewUUID().String()
    wjwt := auth.WJWT{
        UID:   customerId,
        State: "active",
        StandardClaims: jwt.StandardClaims{
            Audience:  "payments",
            ExpiresAt: time.Now().Add(10 * time.Second).Unix(),
            Id:        wuuid.NewUUID().String(),
            IssuedAt:  time.Now().Unix(),
            Issuer:    "auth-service",
        },
    }

    rawRsaPrivateKey, err := readBase64RsaKeyFromFile("testdata/rsa_priv_key_base64_invalid")
    if err != nil {
        return ctx, err
    }

    rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(rawRsaPrivateKey)
    if err != nil {
        return ctx, err
    }

    signedToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, wjwt).SignedString(rsaPrivateKey)
    if err != nil {
        return ctx, err
    }

    return context.WithValue(
        context.WithValue(
            ctx,
            authTokenKey,
            signedToken,
        ),
        customerIdKey,
        customerId,
    ), nil
}

func anUnauthorizedWalleteraCustomer(ctx context.Context) (context.Context, error) {
    customerId := wuuid.NewUUID().String()
    return context.WithValue(
        context.WithValue(
            ctx,
            authTokenKey,
            "",
        ),
        customerIdKey,
        customerId,
    ), nil
}

func anAuthorizedWalleteraCustomer(ctx context.Context) (context.Context, error) {
    customerId := wuuid.NewUUID().String()
    wjwt := auth.WJWT{
        UID:   customerId,
        State: "active",
        StandardClaims: jwt.StandardClaims{
            Audience:  "payments",
            ExpiresAt: time.Now().Add(10 * time.Second).Unix(),
            Id:        wuuid.NewUUID().String(),
            IssuedAt:  time.Now().Unix(),
            Issuer:    "auth-service",
        },
    }

    rawRsaPrivateKey, err := readBase64RsaKeyFromFile("testdata/rsa_priv_key_base64")
    if err != nil {
        return ctx, err
    }

    rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(rawRsaPrivateKey)
    if err != nil {
        return ctx, err
    }

    signedToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, wjwt).SignedString(rsaPrivateKey)
    if err != nil {
        return ctx, err
    }

    return context.WithValue(
        context.WithValue(
            ctx,
            authTokenKey,
            signedToken,
        ),
        customerIdKey,
        customerId,
    ), nil
}

func readBase64RsaKeyFromFile(filePath string) ([]byte, error) {
    base64RsaPrivKey, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    rsaPrivKey, err := base64.StdEncoding.DecodeString(string(base64RsaPrivKey))
    if err != nil {
        return nil, err
    }
    return rsaPrivKey, nil
}

func createMockServerExpectation(ctx context.Context, mockserverExpectation string, ctxKey string) (context.Context, error) {
    if len(mockserverExpectation) == 0 {
        return nil, fmt.Errorf("the mockserver expectation is empty or was not defined")
    }

    rawMockserverExpectation := []byte(mockserverExpectation)

    var unmarshalledExpectation MockServerExpectation
    err := json.Unmarshal(rawMockserverExpectation, &unmarshalledExpectation)
    if err != nil {
        fmt.Errorf("error unmarshalling expectation: %w", err)
    }

    ctx = context.WithValue(ctx, ctxKey, unmarshalledExpectation.ExpectationID)

    err = mockServerClient().CreateExpectation(ctx, rawMockserverExpectation)
    if err != nil {
        fmt.Errorf("error creating mockserver expectations")
    }

    return ctx, nil
}

func createPayment(ctx context.Context, paymentJson string, port int) (api.PostPaymentRes, error) {
    paymentsClient, err := api.NewClient(
        fmt.Sprintf("http://127.0.0.1:%d", port),
        httpauth.NewSecuritySource(authTokenFromCtx(ctx)),
    )
    if err != nil {
        return nil, err
    }
    var payment api.Payment
    err = json.Unmarshal([]byte(paymentJson), &payment)
    if err != nil {
        return nil, fmt.Errorf("failed unmarshalling expected payment: %w", err)
    }
    requestCtx, _ := context.WithTimeout(ctx, 5*time.Second)
    res, err := paymentsClient.PostPayment(requestCtx, &payment, api.PostPaymentParams{})
    if err != nil {
        return nil, err
    }
    return res, nil
}

func thePaymentsServicePublishTheFollowingEvent(ctx context.Context, eventMatcher *godog.DocString) (context.Context, error) {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    expectationId := "matchJSON-" + strconv.Itoa(r.Int())
    ctx, err := createJSONMatcher(ctx, expectationId, eventMatcher.Content)
    if err != nil {
        return ctx, err
    }
    ch := eventsMsgChFromCtx(ctx)
    timeout := time.After(5 * time.Second)
    for {
        select {
        case <-timeout:
            return ctx, fmt.Errorf("timeout waiting for event to be published")
        case msg := <-ch:
            msg.Acknowledger().Ack()
            testLogger.Debug("[TEST] published message", slog.String("message", string(msg.Payload())))
            matched, err := matchJSON(ctx, expectationId, msg.Payload())
            if err != nil {
                testLogger.Debug("[TEST] error matching published event", logattr.Error(err.Error()))
            }
            if matched {
                return ctx, nil
            }
        }
    }
}

func thePaymentsServiceReceiveAPATCHRequestToUpdateThePayment(ctx context.Context, status string) (context.Context, error) {
    payment := paymentCreatedFromCtx(ctx)
    paymentsClient, err := api.NewClient(
        fmt.Sprintf("http://127.0.0.1:%d", privateApiHttpServerPort),
        httpauth.NewSecuritySource(authTokenFromCtx(ctx)),
    )
    if err != nil {
        return nil, err
    }
    requestCtx, _ := context.WithTimeout(ctx, 200*time.Second)
    patchPaymentResp, err := paymentsClient.PatchPayment(requestCtx, &api.PaymentUpdate{
        PaymentId: payment.ID,
        ExternalId: api.OptString{
            Value: wuuid.NewUUID().String(),
            Set:   true,
        },
        Status: api.PaymentStatus(status),
    }, api.PatchPaymentParams{
        XWalleteraCorrelationID: api.OptUUID{
            Value: wuuid.NewUUID(),
            Set:   true,
        },
        PaymentId: payment.ID,
    })
    if err != nil {
        return nil, err
    }
    switch patchPaymentResp.(type) {
    case *api.PatchPaymentOK:
        return ctx, nil
    case *api.PatchPaymentUnauthorized:
        return ctx, fmt.Errorf("unauthorized patch payment request")
    case *api.PatchPaymentInternalServerError:
        return ctx, fmt.Errorf("patch payment request returned internal server error")
    default:
        return ctx, fmt.Errorf("patch payment request return unexpected response")
    }
}

func createJSONMatcher(ctx context.Context, expectationId string, eventMatcher string) (context.Context, error) {
    httpRequestExpectationWrapperTemplate := `
    {
      "id": "%s",
      "httpRequest" : {
        "method": "POST",
        "path": "/matchevent",
        "body": {
            "type": "JSON",
            "json": %s,
            "matchType": "ONLY_MATCHING_FIELDS"
        }
      },
      "httpResponse" : {
        "statusCode" : 201,
        "headers" : {
          "content-type" : [ "application/json" ]
        }
      },
      "priority" : 0,
      "timeToLive" : {
        "unlimited" : true
      },
      "times" : {
        "unlimited" : true
      }
    }
`
    httpRequestExpectationWrapper := fmt.Sprintf(httpRequestExpectationWrapperTemplate, expectationId, eventMatcher)
    return createMockServerExpectation(ctx, httpRequestExpectationWrapper, "")
}

func matchJSON(ctx context.Context, expectationId string, payload []byte) (bool, error) {
    _, err := http.Post(mockserverUrl+"/matchevent", "application/json", bytes.NewReader(payload))
    if err != nil {
        return false, err
    }
    err = verifyExpectationMetWithin(ctx, expectationId, 100*time.Millisecond)
    if err != nil {
        return false, err
    }
    return true, nil
}

const retryPause = 10 * time.Millisecond

func verifyExpectationMetWithin(ctx context.Context, expectationID string, timeout time.Duration) error {
    if retryPause > timeout {
        panic("retryPause is grater than timeout")
    }
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
                time.Sleep(retryPause)
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

func mockServerClient() *msClient.Client {
    mockserverUrl, err := url.Parse(mockserverUrl)
    if err != nil {
        panic("error building mockserver url: " + err.Error())
    }

    return msClient.NewClient(mockserverUrl, http.DefaultClient)
}

func appCtxCancelFuncFromCtx(ctx context.Context) context.CancelFunc {
    return ctx.Value(appCtxCancelFuncKey).(context.CancelFunc)
}

func logsWatcherFromCtx(ctx context.Context) *slogwatcher.Watcher {
    return ctx.Value(logsWatcherKey).(*slogwatcher.Watcher)
}

func authTokenFromCtx(ctx context.Context) string {
    return ctx.Value(authTokenKey).(string)
}

func paymentCreatedFromCtx(ctx context.Context) *api.Payment {
    return ctx.Value(paymentCreatedKey).(*api.Payment)
}

func newZapHandler() (slog.Handler, error) {
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
    zapLogger, err := zapConfig.Build()
    if err != nil {
        return nil, err
    }
    return zapslog.NewHandler(zapLogger.Core()), nil
}
