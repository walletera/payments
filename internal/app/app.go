package app

import (
    "context"
    "encoding/base64"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/golang-jwt/jwt"
    "github.com/walletera/eventskit/eventstoredb"
    "github.com/walletera/eventskit/messages"
    "github.com/walletera/eventskit/rabbitmq"
    "github.com/walletera/payments-types/api"
    paymentevents "github.com/walletera/payments-types/events"
    "github.com/walletera/payments/internal/adapters/input/http/private"
    "github.com/walletera/payments/internal/adapters/input/http/public"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/internal/domain/payment/event/handlers"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/werrors"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
)

const (
    shutdownTimeout                    = 15 * time.Second
    ESDB_ByCategoryProjection_Payments = "$ce-payments.service.payment"
    ESDB_SubscriptionGroupName         = "payments-service"
    PaymentsServiceExchangeName        = "payments.events"
    PaymentServiceExchangeType         = "topic"
    PaymentCreatedRoutingKey           = "payment.created"
)

type App struct {
    rabbitmqHost             string
    rabbitmqPort             int
    rabbitmqUser             string
    rabbitmqPassword         string
    publicAPIHttpServerPort  int
    privateAPIHttpServerPort int
    esdbUrl                  string
    authServiceBase64PubKey  string
    logHandler               slog.Handler
    logger                   *slog.Logger
}

func NewApp(opts ...Option) (*App, error) {
    app := &App{}
    err := setDefaultOpts(app)
    if err != nil {
        return nil, fmt.Errorf("failed setting default options: %w", err)
    }
    for _, opt := range opts {
        opt(app)
    }
    return app, nil
}

func (app *App) Run(ctx context.Context) error {
    appLogger := slog.
        New(app.logHandler).
        With(logattr.ServiceName("payments"))
    app.logger = appLogger

    err := app.execESDBSetupTasks(ctx)
    if err != nil {
        return fmt.Errorf("failed enabling esdb by category projection: %w", err)
    }

    publicApiHttpServer, err := app.startPublicAPIHTTPServer(appLogger)
    if err != nil {
        return fmt.Errorf("failed starting public api http server: %w", err)
    }

    privateApiHttpServer, err := app.startPrivateAPIHTTPServer(appLogger)
    if err != nil {
        return fmt.Errorf("failed starting private api http server: %w", err)
    }

    messageProcessor, err := app.createInternalMessageProcessor(appLogger)
    if err != nil {
        return fmt.Errorf("failed creating internal message processor: %w", err)
    }

    err = messageProcessor.Start(ctx)
    if err != nil {
        return fmt.Errorf("failed starting message processor: %w", err)
    }

    appLogger.Info("payments service started")
    <-ctx.Done()

    shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
    defer cancel()

    app.stopHTTPServers(shutdownCtx, appLogger, publicApiHttpServer, privateApiHttpServer)

    // TODO Close messageProcessor

    appLogger.Info("payments service stopped")
    return nil
}

func (app *App) Stop(ctx context.Context) {
    // TODO implement processor gracefull shutdown
    app.logger.Info("dinopay-gateway stopped")
}

func (app *App) execESDBSetupTasks(ctx context.Context) error {
    err := eventstoredb.EnableByCategoryProjection(ctx, app.esdbUrl)
    if err != nil {
        return fmt.Errorf("failed enabling esdb by category projection: %w", err)
    }

    err = eventstoredb.SetESDBByCategoryProjectionSeparator(ctx, app.esdbUrl)
    if err != nil {
        return fmt.Errorf("failed setting esdb by category projection separator: %w", err)
    }

    err = eventstoredb.CreatePersistentSubscription(app.esdbUrl, ESDB_ByCategoryProjection_Payments, ESDB_SubscriptionGroupName)
    if err != nil {
        return fmt.Errorf("failed creating persistent subscription for %s: %w", ESDB_ByCategoryProjection_Payments, err)
    }

    return nil
}

func (app *App) startPublicAPIHTTPServer(appLogger *slog.Logger) (*http.Server, error) {
    esdbClient, err := eventstoredb.GetESDBClient(app.esdbUrl)
    if err != nil {
        panic(err)
    }
    db := eventstoredb.NewDB(esdbClient)
    paymentService := payment.NewService(db, appLogger)
    securityHandler, err := app.newPublicAPISecurityHandler()
    if err != nil {
        return nil, err
    }
    server, err := api.NewServer(
        public.NewHandler(
            paymentService,
            appLogger.With(logattr.Component("http.PublicAPIHandler")),
        ),
        securityHandler,
    )
    if err != nil {
        panic(err)
    }
    httpServer := &http.Server{
        Addr:    fmt.Sprintf("0.0.0.0:%d", app.publicAPIHttpServerPort),
        Handler: server,
    }

    go func() {
        defer appLogger.Info("http server stopped")
        if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            appLogger.Error("http server error", logattr.Error(err.Error()))
        }
    }()

    appLogger.Info("http server started")

    return httpServer, nil
}

func (app *App) startPrivateAPIHTTPServer(appLogger *slog.Logger) (*http.Server, error) {
    esdbClient, err := eventstoredb.GetESDBClient(app.esdbUrl)
    if err != nil {
        panic(err)
    }
    db := eventstoredb.NewDB(esdbClient)
    paymentService := payment.NewService(db, appLogger)
    securityHandler := private.NewSecurityHandler()
    if err != nil {
        return nil, err
    }
    server, err := api.NewServer(
        private.NewHandler(
            paymentService,
            appLogger.With(logattr.Component("http.PrivateAPIHandler")),
        ),
        securityHandler,
    )
    if err != nil {
        panic(err)
    }
    httpServer := &http.Server{
        Addr:    fmt.Sprintf("0.0.0.0:%d", app.privateAPIHttpServerPort),
        Handler: server,
    }

    go func() {
        defer appLogger.Info("http server stopped")
        if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            appLogger.Error("http server error", logattr.Error(err.Error()))
        }
    }()

    appLogger.Info("http server started")

    return httpServer, nil
}

func (app *App) newPublicAPISecurityHandler() (*public.SecurityHandler, error) {
    pemPubKey, err := base64.StdEncoding.DecodeString(app.authServiceBase64PubKey)
    if err != nil {
        return nil, err
    }
    rsaPubKey, err := jwt.ParseRSAPublicKeyFromPEM(pemPubKey)
    if err != nil {
        return nil, err
    }
    return public.NewSecurityHandler(rsaPubKey), nil
}

func (app *App) stopHTTPServers(ctx context.Context, appLogger *slog.Logger, httpServer ...*http.Server) {
    appLogger.Info("shutting down http server")

    for _, server := range httpServer {
        shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
        defer cancel()
        err := server.Shutdown(shutdownCtx)
        if err != nil {
            appLogger.Error("failed shutting down http server", logattr.Error(err.Error()))
        }
    }
}

func (app *App) createInternalMessageProcessor(logger *slog.Logger) (*messages.Processor[paymentevents.Handler], error) {

    rabbitmqClient, err := rabbitmq.NewClient(
        rabbitmq.WithHost(app.rabbitmqHost),
        rabbitmq.WithPort(uint(app.rabbitmqPort)),
        rabbitmq.WithUser(app.rabbitmqUser),
        rabbitmq.WithPassword(app.rabbitmqPassword),
        rabbitmq.WithExchangeName(PaymentsServiceExchangeName),
        rabbitmq.WithExchangeType(PaymentServiceExchangeType),
        // FIXME this is for publishing events
        // we should not set the consumer routing key
        rabbitmq.WithConsumerRoutingKeys(PaymentCreatedRoutingKey),
    )
    if err != nil {
        return nil, fmt.Errorf("failed creating payments api client: %w", err)
    }

    esdbMessagesConsumer, err := eventstoredb.NewMessagesConsumer(
        app.esdbUrl,
        ESDB_ByCategoryProjection_Payments,
        ESDB_SubscriptionGroupName,
    )
    if err != nil {
        return nil, fmt.Errorf("failed creating esdb messages consumer: %w", err)
    }

    eventsVisitor := handlers.NewPaymentCreatedHandler(rabbitmqClient, logger)

    return messages.NewProcessor[paymentevents.Handler](
            esdbMessagesConsumer,
            paymentevents.NewDeserializer(
                logger.With(logattr.Component("internalEventsDeserializer")),
            ),
            eventsVisitor,
            messages.WithErrorCallback(
                func(processingError werrors.WError) {
                    logger.Error("failed processing esdb event",
                        logattr.Component("payments.service.esdb.MessageProcessor"),
                        logattr.Error(processingError.Message()))
                },
            ),
        ),
        nil
}

func setDefaultOpts(app *App) error {
    zapLogger, err := newZapLogger()
    if err != nil {
        return err
    }
    app.logHandler = zapslog.NewHandler(zapLogger.Core(), nil)
    return nil
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
