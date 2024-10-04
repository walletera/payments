package payment

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/google/uuid"
    "github.com/walletera/message-processor/errors"
    "github.com/walletera/message-processor/eventsourcing"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments-types/events"
)

type Service struct {
    eventDB      eventsourcing.DB
    deserializer *events.Deserializer
}

func NewService(eventDB eventsourcing.DB, logger *slog.Logger) *Service {
    return &Service{
        eventDB:      eventDB,
        deserializer: events.NewDeserializer(logger),
    }
}

func (e *Service) CreatePayment(ctx context.Context, correlationId string, payment api.Payment) error {
    // TODO do some customer related validations
    paymentCreatedEvent := CreatePayment(correlationId, payment)
    err := e.eventDB.AppendEvents(ctx, buildStreamName(paymentCreatedEvent.Data.ID.Value), paymentCreatedEvent)
    if err != nil {
        return fmt.Errorf("failed storing PaymentCreatedEvent: %w", err)
    }
    return nil
}

func (e *Service) UpdatePayment(ctx context.Context, correlationId string, paymentUpdate api.PaymentUpdate) error {
    streamName := buildStreamName(paymentUpdate.PaymentId)
    rawEvents, err := e.eventDB.ReadEvents(ctx, streamName)
    if err != nil {
        return errors.NewInternalError(fmt.Sprintf("failed retrieving events from stream %s: %s", streamName, err.Error()))
    }
    paymentAggregate, err := NewFromEvents(e.deserializer, rawEvents)
    if err != nil {
        return fmt.Errorf("failed creating building payments aggregate from events: %w", err)
    }
    paymentUpdated := paymentAggregate.UpdatePayment(correlationId, UpdateCommand{
        externalId: paymentUpdate.ExternalId,
        status:     paymentUpdate.Status,
    })
    err = e.eventDB.AppendEvents(ctx, buildStreamName(paymentUpdate.PaymentId), paymentUpdated)
    if err != nil {
        return fmt.Errorf("failed storing PaymentUpdatedEvent: %w", err)
    }
    return nil
}

func buildStreamName(aggregateId uuid.UUID) string {
    return fmt.Sprintf("%s.%s", AggregateNamePrefix, aggregateId)
}
