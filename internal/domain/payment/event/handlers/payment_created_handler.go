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
    PaymentCreatedTopic = "payment.events"
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
    return p.publish(ctx, paymentCreatedEvent)
}

func (p *PaymentCreatedHandler) HandlePaymentUpdated(ctx context.Context, paymentUpdated paymentEvents.PaymentUpdated) errors.ProcessingError {
    return p.publish(ctx, paymentUpdated)
}

func (p *PaymentCreatedHandler) publish(ctx context.Context, paymentCreatedEvent events.EventData) errors.ProcessingError {
    err := p.eventPublisher.Publish(ctx, paymentCreatedEvent, PaymentCreatedTopic)
    if err != nil {
        errStr := "failed publishing event"
        p.logger.Error(errStr, logattr.EventType(paymentCreatedEvent.Type()), logattr.Error(err.Error()))
        return errors.NewInternalError(fmt.Sprintf("%s: %s", errStr, err.Error()))
    }
    return nil
}
