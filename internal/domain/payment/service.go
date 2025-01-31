package payment

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/google/uuid"
    "github.com/walletera/eventskit/eventsourcing"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments-types/events"
    "github.com/walletera/werrors"
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

func (e *Service) CreatePayment(
    ctx context.Context,
    correlationId string,
    customerId uuid.UUID,
    payment api.Payment,
) (*events.PaymentCreated, werrors.WError) {
    paymentCreatedEvent, err := CreatePayment(
        correlationId,
        customerId,
        payment,
    )
    if err != nil {
        return nil, werrors.NewWrappedError(err, "failed creating payment")
    }
    streamName := buildStreamName(paymentCreatedEvent.Data.ID)
    err = e.eventDB.AppendEvents(
        ctx,
        streamName,
        eventsourcing.ExpectedAggregateVersion{IsNew: true},
        paymentCreatedEvent,
    )
    if err != nil {
        return nil, werrors.NewWrappedError(err, "failed appending event %s to the stream %s", paymentCreatedEvent.EventType, streamName)
    }
    return &paymentCreatedEvent, nil
}

func (e *Service) UpdatePayment(ctx context.Context, correlationId string, paymentUpdate *api.PaymentUpdate) werrors.WError {
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
    streamName := buildStreamName(paymentUpdate.PaymentId)
    err = e.eventDB.AppendEvents(
        ctx,
        streamName,
        eventsourcing.ExpectedAggregateVersion{Version: paymentAggregate.Version()},
        paymentUpdated,
    )
    if err != nil {
        return werrors.NewWrappedError(err, "failed appending event %s to the stream %s", paymentUpdated.EventType, streamName)
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

func (e *Service) buildAggregateFromStoredEvents(ctx context.Context, paymentId uuid.UUID) (*Aggregate, werrors.WError) {
    streamName := buildStreamName(paymentId)
    rawEvents, err := e.eventDB.ReadEvents(ctx, streamName)
    if err != nil {
        return nil, werrors.NewWrappedError(err, "failed retrieving events from stream %s: %s", streamName)
    }
    paymentAggregate, err := NewFromEvents(e.deserializer, rawEvents)
    if err != nil {
        return nil, werrors.NewWrappedError(err, "failed creating building payments aggregate from events")
    }
    return paymentAggregate, nil
}

func buildStreamName(aggregateId uuid.UUID) string {
    return fmt.Sprintf("%s.%s", AggregateNamePrefix, aggregateId)
}
