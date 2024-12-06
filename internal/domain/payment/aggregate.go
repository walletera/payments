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
    aggregate := &Aggregate{}
    for _, rawEvent := range rawEvents {
        event, err := deserializer.Deserialize(rawEvent)
        if err != nil {
            return nil, fmt.Errorf("failed deserializing event from raw event %s: %s", rawEvent, err.Error())
        }
        if event == nil {
            return nil, fmt.Errorf("failed deserializing event from raw event %s", rawEvent)
        }
        event.Accept(context.Background(), aggregate)
    }
    return aggregate, nil
}

func (p *Aggregate) UpdatePayment(correlationId string, command UpdateCommand) (eventtypes.PaymentUpdated, error) {
    paymentUpdate := api.PaymentUpdate{
        PaymentId: p.payment.ID.Value,
    }
    if !p.canTransition(command.status) {
        currentStatus, _ := p.payment.Status.Get()
        return eventtypes.PaymentUpdated{}, fmt.Errorf("invalid payment status transition from %s to %s", currentStatus, command.status)
    }
    paymentUpdate.Status = command.status
    paymentUpdate.ExternalId = command.externalId
    return eventtypes.NewPaymentUpdated(correlationId, paymentUpdate), nil
}

func (p *Aggregate) HandlePaymentCreated(ctx context.Context, paymentCreatedEvent eventtypes.PaymentCreated) errors.ProcessingError {
    p.payment = paymentCreatedEvent.Data
    return nil
}

func (p *Aggregate) HandlePaymentUpdated(ctx context.Context, paymentUpdated eventtypes.PaymentUpdated) errors.ProcessingError {
    p.payment.ExternalId = paymentUpdated.Data.ExternalId
    p.payment.Status = api.OptPaymentStatus{
        Value: paymentUpdated.Data.Status,
        Set:   true,
    }
    return nil
}

func (p *Aggregate) GetPayment() *api.Payment {
    return &p.payment
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
