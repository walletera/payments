package public

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/google/uuid"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/adapters/input/http/shared"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
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
    _, err := getCustomerIdFromCtx(ctx)
    if err != nil {
        h.logger.Warn(err.Error())
        return &api.GetPaymentUnauthorized{}, nil
    }
    payment, err := h.service.GetPayment(ctx, params.PaymentId)
    if err != nil {
        // FIXME improve error handling
        h.logger.Error("failed getting payment", logattr.Error(err.Error()))
        return &api.GetPaymentNotFound{}, nil
    }
    return payment, nil
}

func (h *Handler) PatchPayment(ctx context.Context, req *api.PaymentUpdate, params api.PatchPaymentParams) (api.PatchPaymentRes, error) {
    return &api.PatchPaymentUnauthorized{}, nil
}

func (h *Handler) PostPayment(ctx context.Context, req *api.Payment, params api.PostPaymentParams) (api.PostPaymentRes, error) {
    var correlationId string
    if params.XWalleteraCorrelationID.Set {
        correlationId = params.XWalleteraCorrelationID.Value.String()
    } else {
        correlationId = wuuid.NewUUID().String()
    }
    customerId, err := getCustomerIdFromCtx(ctx)
    if err != nil {
        h.logger.Warn(err.Error())
        return &api.PostPaymentUnauthorized{}, nil
    }
    parsedCustomerId, err := uuid.Parse(customerId)
    if err != nil {
        errorCode := wuuid.NewUUID()
        h.logger.Error(
            "error parsing customerId: "+err.Error(),
            logattr.ErrorCode(errorCode),
            logattr.CorrelationId(correlationId),
        )
        return &api.PostPaymentInternalServerError{
            ErrorMessage: "unexpected internal error",
            ErrorCode:    errorCode,
        }, nil
    }
    return h.sharedOperations.CreatePayment(ctx, correlationId, parsedCustomerId, *req)
}

func getCustomerIdFromCtx(ctx context.Context) (string, error) {
    customerIdFromCtx := ctx.Value(WJWTCustomerIdCtxKey)
    if customerIdFromCtx == nil {
        return "", fmt.Errorf("customerId not found in context")
    }
    customerId, _ := customerIdFromCtx.(string)
    if len(customerId) == 0 {
        return "", fmt.Errorf("customerId is empty")
    }
    return customerId, nil
}
