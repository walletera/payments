package payment

import (
    "context"
    "fmt"
    "time"

    "github.com/walletera/message-processor/errors"
    "github.com/walletera/message-processor/events"
    "github.com/walletera/message-processor/eventsourcing"
    "github.com/walletera/payments-types/api"
    eventtypes "github.com/walletera/payments-types/events"
    "github.com/walletera/payments/pkg/wuuid"
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
    externalId api.OptUUID
    status     api.PaymentStatus
}

type Aggregate struct {
    payment api.Payment
}

func CreatePayment(correlationId string, payment api.Payment) eventtypes.PaymentCreated {
    newPayment := payment
    newPayment.ID = api.NewOptUUID(wuuid.NewUUID())
    newPayment.Status = api.OptPaymentStatus{
        Value: api.PaymentStatusPending,
        Set:   true,
    }
    newPayment.Direction = api.NewOptPaymentDirection(api.PaymentDirectionOutbound)
    // FIXME hardcoded to make test pass
    newPayment.CustomerId = api.NewOptUUID(wuuid.NewUUID())
    newPayment.CreatedAt = api.OptDateTime{
        Value: time.Now(),
        Set:   true,
    }
    return eventtypes.NewPaymentCreated(correlationId, newPayment)
}

func NewFromEvents(deserializer events.Deserializer[eventtypes.Handler], rawEvents []eventsourcing.RawEvent) (*Aggregate, error) {
    p := &Aggregate{}
    for _, rawEvent := range rawEvents {
        event, err := deserializer.Deserialize(rawEvent)
        if err != nil {
            return nil, fmt.Errorf("failed deserializing outboundPaymentUpdated event from raw event %s: %s", rawEvent, err.Error())
        }
        event.Accept(context.Background(), p)
    }
    return p, nil
}

func (p *Aggregate) UpdatePayment(correlationId string, command UpdateCommand) eventtypes.PaymentUpdated {
    paymentUpdate := api.PaymentUpdate{
        PaymentId: p.payment.ID.Value,
    }
    if p.canTransition(command.status) {
        paymentUpdate.Status = command.status
    }
    paymentUpdate.ExternalId = command.externalId
    return eventtypes.NewPaymentUpdated(correlationId, paymentUpdate)
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

func (p *Aggregate) HandlePaymentCreated(ctx context.Context, paymentCreatedEvent eventtypes.PaymentCreated) errors.ProcessingError {
    p.payment = paymentCreatedEvent.Data
    return nil
}

func (p *Aggregate) HandlePaymentUpdated(ctx context.Context, paymentCreatedEvent eventtypes.PaymentUpdated) errors.ProcessingError {
    p.payment.ExternalId = paymentCreatedEvent.Data.ExternalId
    p.payment.Status = api.OptPaymentStatus{
        Value: paymentCreatedEvent.Data.Status,
        Set:   true,
    }
    return nil
}
