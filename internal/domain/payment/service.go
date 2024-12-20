package payment

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/google/uuid"
    "github.com/walletera/eventskit/eventsourcing"
    "github.com/walletera/message-processor/errors"
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

func (e *Service) CreatePayment(ctx context.Context, correlationId string, payment api.Payment) (*events.PaymentCreated, error) {
    // TODO do some customer related validations
    paymentCreatedEvent := CreatePayment(correlationId, payment)
    streamName := buildStreamName(paymentCreatedEvent.Data.ID)
    err := e.eventDB.AppendEvents(
        ctx,
        streamName,
        eventsourcing.ExpectedAggregateVersion{IsNew: true},
        paymentCreatedEvent,
    )
    if err != nil {
        return nil, fmt.Errorf("failed appending event %s to the stream %s: %w", paymentCreatedEvent.EventType, streamName, err)
    }
    return &paymentCreatedEvent, nil
}

func (e *Service) UpdatePayment(ctx context.Context, correlationId string, paymentUpdate *api.PaymentUpdate) error {
    paymentAggregate, err := e.buildAggregateFromStoredEvents(ctx, paymentUpdate.PaymentId)
    if err != nil {
        return err
    }
    paymentUpdated, err := paymentAggregate.UpdatePayment(correlationId, UpdateCommand{
        externalId: paymentUpdate.ExternalId,
        status:     paymentUpdate.Status,
    })
    if err != nil {
        return err
    }
    err = e.eventDB.AppendEvents(
        ctx,
        buildStreamName(paymentUpdate.PaymentId),
        eventsourcing.ExpectedAggregateVersion{Version: paymentAggregate.Version()},
        paymentUpdated,
    )
    if err != nil {
        return fmt.Errorf("failed storing PaymentUpdatedEvent: %w", err)
    }
    return nil
}

func (e *Service) GetPayment(ctx context.Context, paymentId uuid.UUID) (*api.Payment, error) {
    paymentAggregate, err := e.buildAggregateFromStoredEvents(ctx, paymentId)
    if err != nil {
        return nil, err
    }
    return paymentAggregate.Payment(), nil
}

func (e *Service) buildAggregateFromStoredEvents(ctx context.Context, paymentId uuid.UUID) (*Aggregate, error) {
    streamName := buildStreamName(paymentId)
    rawEvents, err := e.eventDB.ReadEvents(ctx, streamName)
    if err != nil {
        return nil, errors.NewInternalError(fmt.Sprintf("failed retrieving events from stream %s: %s", streamName, err.Error()))
    }
    paymentAggregate, err := NewFromEvents(e.deserializer, rawEvents)
    if err != nil {
        return nil, fmt.Errorf("failed creating building payments aggregate from events: %w", err)
    }
    return paymentAggregate, nil
}

func buildStreamName(aggregateId uuid.UUID) string {
    return fmt.Sprintf("%s.%s", AggregateNamePrefix, aggregateId)
}
