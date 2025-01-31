package logattr

import (
    "log/slog"

    "github.com/google/uuid"
)

func ServiceName(serviceName string) slog.Attr {
    return slog.String("service_name", serviceName)
}

func Component(component string) slog.Attr {
    return slog.String("component", component)
}

func PaymentId(withdrawalId string) slog.Attr {
    return slog.String("payment_id", withdrawalId)
}

func EventType(eventType string) slog.Attr {
    return slog.String("event_type", eventType)
}

func Error(err string) slog.Attr {
    return slog.String("error", err)
}

func ErrorCode(err uuid.UUID) slog.Attr {
    return slog.String("error_code", err.String())
}

func CorrelationId(correlationId string) slog.Attr {
    return slog.String("correlation_id", correlationId)
}
