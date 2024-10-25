package events

import (
    "encoding/json"
    "fmt"
    "log"
    "log/slog"

    "github.com/walletera/message-processor/events"
    "github.com/walletera/payments-types/api"
)

// interface compliance verification
var _ events.Deserializer[Handler] = (*Deserializer)(nil)

type Deserializer struct {
    logger *slog.Logger
}

func NewDeserializer(logger *slog.Logger) *Deserializer {
    return &Deserializer{logger: logger}
}

func (d *Deserializer) Deserialize(rawPayload []byte) (events.Event[Handler], error) {
    var event events.EventEnvelope
    err := json.Unmarshal(rawPayload, &event)
    if err != nil {
        return nil, fmt.Errorf("error deserializing message with payload %s: %w", rawPayload, err)
    }
    switch event.Type {
    case "PaymentCreated":
        var payment api.Payment
        err := json.Unmarshal(event.Data, &payment)
        if err != nil {
            log.Printf("error deserializing PaymentCreated event data %s: %s", event.Data, err.Error())
        }
        return NewPaymentCreated(event.CorrelationID, payment), nil
    case "PaymentUpdated":
        var payment api.PaymentUpdate
        err := json.Unmarshal(event.Data, &payment)
        if err != nil {
            log.Printf("error deserializing PaymentUpdated event data %s: %s", event.Data, err.Error())
        }
        return NewPaymentUpdated(event.CorrelationID, payment), nil
    default:
        d.logger.Warn("unexpected event type", slog.String("eventType", event.Type))
        return nil, nil
    }
}