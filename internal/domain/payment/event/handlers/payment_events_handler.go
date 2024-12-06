package handlers

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/walletera/message-processor/errors"
    "github.com/walletera/message-processor/events"
    paymentEvents "github.com/walletera/payments-types/events"
    "github.com/walletera/payments/pkg/logattr"
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

func (p *PaymentCreatedHandler) HandlePaymentCreated(ctx context.Context, paymentCreatedEvent paymentEvents.PaymentCreated) errors.ProcessingError {
    return p.publish(ctx, paymentCreatedEvent, events.RoutingInfo{
        Topic:      PaymentsTopic,
        RoutingKey: PaymentCreatedRoutingKey,
    })
}

func (p *PaymentCreatedHandler) HandlePaymentUpdated(ctx context.Context, paymentUpdated paymentEvents.PaymentUpdated) errors.ProcessingError {
    return p.publish(ctx, paymentUpdated, events.RoutingInfo{
        Topic:      PaymentsTopic,
        RoutingKey: PaymentUpdatedRoutingKey,
    })
}

func (p *PaymentCreatedHandler) publish(ctx context.Context, eventData events.EventData, routingInfo events.RoutingInfo) errors.ProcessingError {
    err := p.eventPublisher.Publish(ctx, eventData, routingInfo)
    if err != nil {
        errStr := "failed publishing event"
        p.logger.Error(errStr,
            logattr.CorrelationId(eventData.CorrelationID()),
            logattr.EventType(eventData.Type()),
            logattr.Error(err.Error()),
        )
        return errors.NewInternalError(fmt.Sprintf("%s: %s", errStr, err.Error()))
    }
    p.logger.Info("event published",
        logattr.CorrelationId(eventData.CorrelationID()),
        logattr.EventType(eventData.Type()),
    )
    return nil
}
