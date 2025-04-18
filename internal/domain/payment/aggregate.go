package payment

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/walletera/eventskit/events"
    "github.com/walletera/eventskit/eventsourcing"
    "github.com/walletera/payments-types/api"
    eventtypes "github.com/walletera/payments-types/events"
    "github.com/walletera/werrors"
)

const (
    AggregateNamePrefix = "payments.service.payment"
)

var validStatusTransitions = map[api.PaymentStatus][]api.PaymentStatus{
    api.PaymentStatusPending: {
        api.PaymentStatusDelivered,
        api.PaymentStatusConfirmed,
        api.PaymentStatusFailed,
    },
}

type UpdateCommand struct {
    externalId api.OptString
    status     api.PaymentStatus
}

type Aggregate struct {
    payment api.Payment
    version uint64
}

func CreatePayment(correlationId string, customerId uuid.UUID, payment api.Payment) (eventtypes.PaymentCreated, werrors.WError) {
    newPayment := payment
    newPayment.CustomerId = api.OptUUID{
        Value: customerId,
        Set:   true,
    }
    newPayment.Status = api.OptPaymentStatus{
        Value: api.PaymentStatusPending,
        Set:   true,
    }
    newPayment.Direction = api.NewOptPaymentDirection(api.PaymentDirectionOutbound)
    newPayment.CreatedAt = api.OptDateTime{
        Value: time.Now(),
        Set:   true,
    }
    return eventtypes.NewPaymentCreated(correlationId, newPayment), nil
}

func NewFromEvents(deserializer events.Deserializer[eventtypes.Handler], retrievedEvents []eventsourcing.RetrievedEvent) (*Aggregate, werrors.WError) {
    aggregate := &Aggregate{}
    for _, retrievedEvent := range retrievedEvents {
        event, err := deserializer.Deserialize(retrievedEvent.RawEvent)
        if err != nil {
            return nil, werrors.NewNonRetryableInternalError("failed deserializing event from raw event %s: %s", retrievedEvent.RawEvent, err.Error())
        }
        if event == nil {
            return nil, werrors.NewNonRetryableInternalError("failed deserializing event from raw event %s", retrievedEvent.RawEvent)
        }
        event.Accept(context.Background(), aggregate)
        aggregate.version = retrievedEvent.AggregateVersion
    }
    return aggregate, nil
}

func (p *Aggregate) UpdatePayment(correlationId string, command UpdateCommand) (eventtypes.PaymentUpdated, werrors.WError) {
    paymentUpdate := api.PaymentUpdate{
        PaymentId: p.payment.ID,
    }
    if !p.canTransition(command.status) {
        currentStatus, _ := p.payment.Status.Get()
        return eventtypes.PaymentUpdated{}, werrors.NewValidationError(
            "invalid payment status transition from %s to %s",
            currentStatus,
            command.status,
        )
    }
    paymentUpdate.Status = command.status
    paymentUpdate.ExternalId = command.externalId
    return eventtypes.NewPaymentUpdated(correlationId, paymentUpdate), nil
}

func (p *Aggregate) HandlePaymentCreated(ctx context.Context, paymentCreatedEvent eventtypes.PaymentCreated) werrors.WError {
    p.payment = paymentCreatedEvent.Data
    return nil
}

func (p *Aggregate) HandlePaymentUpdated(ctx context.Context, paymentUpdated eventtypes.PaymentUpdated) werrors.WError {
    p.payment.ExternalId = paymentUpdated.Data.ExternalId
    p.payment.Status = api.OptPaymentStatus{
        Value: paymentUpdated.Data.Status,
        Set:   true,
    }
    return nil
}

func (p *Aggregate) Payment() *api.Payment {
    return &p.payment
}

func (a *Aggregate) Version() uint64 {
    return a.version
}

func (p *Aggregate) canTransition(status api.PaymentStatus) bool {
    currentStatus, ok := p.payment.Status.Get()
    if !ok {
        // this should not happen
        // TODO log a warning
        return false
    }
    validTransitions, ok := validStatusTransitions[currentStatus]
    if !ok {
        return false
    }
    for _, transition := range validTransitions {
        if transition == status {
            return true
        }
    }
    return false
}
