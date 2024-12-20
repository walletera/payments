package handlers

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/walletera/eventskit/events"
    paymentEvents "github.com/walletera/payments-types/events"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/werrors"
)

const (
    PaymentsTopic            = "payments.events"
    PaymentCreatedRoutingKey = "payment.created"
    PaymentUpdatedRoutingKey = "payment.updated"
)

type PaymentCreatedHandler struct {
    eventPublisher events.Publisher
    logger         *slog.Logger
}

func NewPaymentCreatedHandler(eventPublisher events.Publisher, logger *slog.Logger) *PaymentCreatedHandler {
    return &PaymentCreatedHandler{
        eventPublisher: eventPublisher,
        logger:         logger,
    }
}

func (p *PaymentCreatedHandler) HandlePaymentCreated(ctx context.Context, paymentCreatedEvent paymentEvents.PaymentCreated) werrors.WError {
    return p.publish(ctx, paymentCreatedEvent, events.RoutingInfo{
        Topic:      PaymentsTopic,
        RoutingKey: PaymentCreatedRoutingKey,
    })
}

func (p *PaymentCreatedHandler) HandlePaymentUpdated(ctx context.Context, paymentUpdated paymentEvents.PaymentUpdated) werrors.WError {
    return p.publish(ctx, paymentUpdated, events.RoutingInfo{
        Topic:      PaymentsTopic,
        RoutingKey: PaymentUpdatedRoutingKey,
    })
}

func (p *PaymentCreatedHandler) publish(ctx context.Context, eventData events.EventData, routingInfo events.RoutingInfo) werrors.WError {
    err := p.eventPublisher.Publish(ctx, eventData, routingInfo)
    if err != nil {
        errStr := "failed publishing event"
        p.logger.Error(errStr,
            logattr.CorrelationId(eventData.CorrelationID()),
            logattr.EventType(eventData.Type()),
            logattr.Error(err.Error()),
        )
        return werrors.NewRetryableInternalError(fmt.Sprintf("%s: %s", errStr, err.Error()))
    }
    p.logger.Info("event published",
        logattr.CorrelationId(eventData.CorrelationID()),
        logattr.EventType(eventData.Type()),
    )
    return nil
}
