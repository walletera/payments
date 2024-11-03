package http

import (
    "context"
    "log/slog"

    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
)

type Handler struct {
    service *payment.Service
    logger  *slog.Logger
}

var _ api.Handler = (*Handler)(nil)

func NewHandler(service *payment.Service, logger *slog.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

func (h *Handler) PatchPayment(ctx context.Context, req *api.PaymentUpdate, params api.PatchPaymentParams) (api.PatchPaymentRes, error) {
    //TODO implement me
    panic("implement me")
}

func (h *Handler) PostPayment(ctx context.Context, req *api.Payment, _ api.PostPaymentParams) (api.PostPaymentRes, error) {
    correlationId := wuuid.NewUUID()
    err := h.service.CreatePayment(ctx, correlationId.String(), *req)
    if err != nil {
        h.logger.Error("payment creation failed", logattr.Error(err.Error()))
        return &api.PostPaymentInternalServerError{}, nil
    }
    h.logger.Info("payment created")
    return &api.PostPaymentCreated{}, nil
}
