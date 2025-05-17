package payment

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/google/uuid"
    "github.com/walletera/eventskit/eventsourcing"
    "github.com/walletera/payments-types/events"
    privapi "github.com/walletera/payments-types/privateapi"
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
    payment privapi.Payment,
) (events.PaymentCreated, werrors.WError) {
    payment.CreatedAt = time.Now()
    err := payment.Validate()
    if err != nil {
        return events.PaymentCreated{}, werrors.NewValidationError(err.Error())
    }
    paymentCreatedEvent := events.NewPaymentCreated(correlationId, payment)
    streamName := buildStreamName(paymentCreatedEvent.Data.ID)
    appendErr := e.eventDB.AppendEvents(
        ctx,
        streamName,
        eventsourcing.ExpectedAggregateVersion{IsNew: true},
        paymentCreatedEvent,
    )
    if appendErr != nil {
        return events.PaymentCreated{}, werrors.NewWrappedError(
            appendErr,
            "failed appending event %s to the stream %s",
            paymentCreatedEvent.EventType,
            streamName,
        )
    }
    return paymentCreatedEvent, nil
}

func (e *Service) UpdatePayment(ctx context.Context, correlationId string, paymentUpdate privapi.PaymentUpdate) werrors.WError {
    paymentAggregate, err := e.buildAggregateFromStoredEvents(ctx, paymentUpdate.PaymentId)
    if err != nil {
        return err
    }
    paymentUpdated, err := paymentAggregate.UpdatePayment(correlationId, paymentUpdate)
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

func (e *Service) GetPayment(ctx context.Context, paymentId uuid.UUID) (privapi.Payment, werrors.WError) {
    paymentAggregate, err := e.buildAggregateFromStoredEvents(ctx, paymentId)
    if err != nil {
        return privapi.Payment{}, err
    }
    return paymentAggregate.Payment(), nil
}

func (e *Service) buildAggregateFromStoredEvents(ctx context.Context, paymentId uuid.UUID) (*Aggregate, werrors.WError) {
    streamName := buildStreamName(paymentId)
    rawEvents, err := e.eventDB.ReadEvents(ctx, streamName)
    if err != nil {
        return nil, werrors.NewWrappedError(err, "failed retrieving events from stream %s", streamName)
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
