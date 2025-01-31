package private

import (
    "context"
    "log/slog"

    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/adapters/input/http/shared"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
    "github.com/walletera/werrors"
)

type Handler struct {
    sharedOperations *shared.Operations
    service          *payment.Service
    logger           *slog.Logger
}

var _ api.Handler = (*Handler)(nil)

func NewHandler(service *payment.Service, logger *slog.Logger) *Handler {
    return &Handler{
        sharedOperations: shared.NewOperations(service, logger),
        service:          service,
        logger:           logger,
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
        errorCode := wuuid.NewUUID()
        h.logger.Error(
            "payment creation failed",
            logattr.CorrelationId(correlationId),
            logattr.Error(err.Error()),
            logattr.ErrorCode(errorCode),
        )
        switch err.Code() {
        case werrors.ValidationErrorCode:
            return &api.PatchPaymentBadRequest{
                ErrorMessage: err.Message(),
                ErrorCode:    errorCode,
            }, nil
        default:
            return &api.PatchPaymentInternalServerError{
                ErrorMessage: "unexpected internal error",
                ErrorCode:    errorCode,
            }, nil
        }
    }
    return &api.PatchPaymentOK{}, nil
}

func (h *Handler) PostPayment(ctx context.Context, req *api.Payment, params api.PostPaymentParams) (api.PostPaymentRes, error) {
    var correlationId string
    if params.XWalleteraCorrelationID.Set {
        correlationId = params.XWalleteraCorrelationID.Value.String()
    } else {
        correlationId = wuuid.NewUUID().String()
    }
    if !req.CustomerId.Set {
        errorCode := wuuid.NewUUID()
        h.logger.Error(
            "missing customerId in request",
            logattr.ErrorCode(errorCode),
            logattr.CorrelationId(correlationId),
        )
        resp := api.PostPaymentBadRequest{
            ErrorMessage: "missing customerId",
            ErrorCode:    errorCode,
        }
        return &resp, nil
    }
    return h.sharedOperations.CreatePayment(ctx, correlationId, req.CustomerId.Value, *req)
}
