package app

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    procerrors "github.com/walletera/message-processor/errors"
    "github.com/walletera/message-processor/eventstoredb"
    "github.com/walletera/message-processor/messages"
    "github.com/walletera/message-processor/rabbitmq"
    "github.com/walletera/payments-types/api"
    paymentevents "github.com/walletera/payments-types/events"
    httpadapter "github.com/walletera/payments/internal/adapters/input/http"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/internal/domain/payment/event/handlers"
    "github.com/walletera/payments/pkg/logattr"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
)

const (
    shutdownTimeout                    = 15 * time.Second
    ESDB_ByCategoryProjection_Payments = "$ce-payments.service.payment"
    ESDB_SubscriptionGroupName         = "payments-service"
    PaymentsServiceExchangeName        = "paymentsService.events"
    PaymentServiceExchangeType         = "topic"
)

type App struct {
    rabbitmqUrl    string
    httpServerPort int
    esdbUrl        string
    logHandler     slog.Handler
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

    httpServer, err := app.startHTTPServer(appLogger)
    if err != nil {
        return fmt.Errorf("failed starting HTTP server: %w", err)
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

    app.stopHTTPServer(shutdownCtx, httpServer, appLogger)

    // TODO Close messageProcessor

    appLogger.Info("payments service stopped")
    return nil
}

func (app *App) startHTTPServer(appLogger *slog.Logger) (*http.Server, error) {
    esdbClient, err := eventstoredb.GetESDBClient(app.esdbUrl)
    if err != nil {
        panic(err)
    }
    db := eventstoredb.NewDB(esdbClient)
    paymentService := payment.NewService(db, appLogger)
    server, err := api.NewServer(
        httpadapter.NewHandler(
            paymentService,
            appLogger.With(logattr.Component("http-handler")),
        ),
    )
    if err != nil {
        panic(err)
    }
    httpServer := &http.Server{
        Addr:    fmt.Sprintf("0.0.0.0:%d", app.httpServerPort),
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

func (app *App) stopHTTPServer(ctx context.Context, httpServer *http.Server, appLogger *slog.Logger) {
    appLogger.Info("shutting down http server")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
    defer cancel()
    err := httpServer.Shutdown(shutdownCtx)
    if err != nil {
        appLogger.Error("failed shutting down http server", logattr.Error(err.Error()))
    }
}

func (app *App) createInternalMessageProcessor(logger *slog.Logger) (*messages.Processor[paymentevents.Handler], error) {

    rabbitmqClient, err := rabbitmq.NewClient(
        rabbitmq.WithExchangeName(PaymentsServiceExchangeName),
        rabbitmq.WithExchangeType(PaymentServiceExchangeType),
        rabbitmq.WithConsumerRoutingKeys(handlers.PaymentCreatedTopic),
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
                func(processingError procerrors.ProcessingError) {
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
