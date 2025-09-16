package payment

import (
    "context"

    "github.com/walletera/eventskit/events"
    "github.com/walletera/eventskit/eventsourcing"
    paymentevents "github.com/walletera/payments-types/events"
    privapi "github.com/walletera/payments-types/privateapi"
    "github.com/walletera/werrors"
)

const (
    AggregateNamePrefix = "payments.service.payment"
)

var validStatusTransitions = map[privapi.PaymentStatus][]privapi.PaymentStatus{
    privapi.PaymentStatusPending: {
        privapi.PaymentStatusDelivered,
        privapi.PaymentStatusConfirmed,
        privapi.PaymentStatusFailed,
    },
}

type Aggregate struct {
    payment privapi.Payment
    version uint64
}

func NewFromEvents(deserializer events.Deserializer[paymentevents.Handler], retrievedEvents []eventsourcing.RetrievedEvent) (*Aggregate, werrors.WError) {
    aggregate := &Aggregate{}
    for _, retrievedEvent := range retrievedEvents {
        event, err := deserializer.Deserialize(retrievedEvent.RawEvent)
        if err != nil {
            return nil, werrors.NewNonRetryableInternalError("failed deserializing event from raw event %s: %s", retrievedEvent.RawEvent, err.Error())
        }
        if event == nil {
            return nil, werrors.NewNonRetryableInternalError("failed deserializing event from raw event %s", retrievedEvent.RawEvent)
        }
        acceptErr := event.Accept(context.Background(), aggregate)
        if acceptErr != nil {
            return nil, werrors.NewNonRetryableInternalError("failed accepting event: %s", acceptErr.Error())
        }
        aggregate.version = retrievedEvent.AggregateVersion
    }
    return aggregate, nil
}

func (a *Aggregate) UpdatePayment(correlationId string, paymentUpdate privapi.PaymentUpdate) (paymentevents.PaymentUpdated, werrors.WError) {
    if !a.canTransition(paymentUpdate.Status) {
        currentStatus := a.payment.Status
        return paymentevents.PaymentUpdated{}, werrors.NewValidationError(
            "invalid payment status transition from %s to %s",
            currentStatus,
            paymentUpdate.Status,
        )
    }
    return paymentevents.NewPaymentUpdated(a.version+1, correlationId, paymentUpdate), nil
}

func (a *Aggregate) HandlePaymentCreated(ctx context.Context, paymentCreatedEvent paymentevents.PaymentCreated) werrors.WError {
    a.payment = paymentCreatedEvent.Data
    return nil
}

func (a *Aggregate) HandlePaymentUpdated(ctx context.Context, paymentUpdated paymentevents.PaymentUpdated) werrors.WError {
    a.payment.ExternalId = paymentUpdated.Data.ExternalId
    a.payment.Status = paymentUpdated.Data.Status
    return nil
}

func (a *Aggregate) Payment() privapi.Payment {
    return a.payment
}

func (a *Aggregate) Version() uint64 {
    return a.version
}

func (a *Aggregate) canTransition(status privapi.PaymentStatus) bool {
    currentStatus := a.payment.Status
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
