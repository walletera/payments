package http

import (
    "context"
    "log/slog"

    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
    "github.com/walletera/werrors"
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

func (h *Handler) GetPayment(ctx context.Context, params api.GetPaymentParams) (api.GetPaymentRes, error) {
    payment, err := h.service.GetPayment(ctx, params.PaymentId)
    if err != nil {
        // FIXME improve error handling
        h.logger.Error("failed getting payment", logattr.Error(err.Error()))
        return &api.GetPaymentNotFound{}, nil
    }
    return payment, nil
}

func (h *Handler) PatchPayment(ctx context.Context, req *api.PaymentUpdate, params api.PatchPaymentParams) (api.PatchPaymentRes, error) {
    var correlationId string
    if params.XWalleteraCorrelationID.Set {
        correlationId = params.XWalleteraCorrelationID.Value.String()
    } else {
        correlationId = wuuid.NewUUID().String()
    }
    err := h.service.UpdatePayment(ctx, correlationId, req)
    if err != nil {
        h.logger.Error("payment creation failed", logattr.Error(err.Error()))
        switch err.Code() {
        case werrors.ValidationErrorCode:
            resp := api.ErrorMessage(err.Message())
            return &resp, nil
        default:
            return &api.PatchPaymentInternalServerError{}, nil
        }
    }
    return &api.PatchPaymentOK{}, nil
}

func (h *Handler) PostPayment(ctx context.Context, req *api.Payment, _ api.PostPaymentParams) (api.PostPaymentRes, error) {
    correlationId := wuuid.NewUUID()
    paymentCreated, err := h.service.CreatePayment(ctx, correlationId.String(), *req)
    if err != nil {
        h.logger.Error("payment creation failed", logattr.Error(err.Error()))
        switch err.Code() {
        case werrors.ResourceAlreadyExistErrorCode:
            resp := api.PostPaymentConflict("the payment you are trying to create already exist")
            return &resp, nil
        case werrors.ValidationErrorCode:
            resp := api.PostPaymentBadRequest(err.Message())
            return &resp, nil
        default:
            return &api.PostPaymentInternalServerError{}, nil
        }
    }
    h.logger.Info("payment created",
        logattr.CorrelationId(correlationId.String()),
        logattr.PaymentId(paymentCreated.Data.ID.String()),
    )
    return &paymentCreated.Data, nil
}
