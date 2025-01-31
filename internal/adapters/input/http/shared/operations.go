package shared

import (
    "context"
    "log/slog"

    "github.com/google/uuid"
    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
    "github.com/walletera/werrors"
)

type Operations struct {
    service *payment.Service
    logger  *slog.Logger
}

func NewOperations(service *payment.Service, logger *slog.Logger) *Operations {
    return &Operations{
        service: service,
        logger:  logger,
    }
}

func (h *Operations) CreatePayment(
    ctx context.Context,
    correlationId string,
    customerId uuid.UUID,
    payment api.Payment,
) (api.PostPaymentRes, error) {

    paymentCreated, createPaymentErr := h.service.CreatePayment(ctx, correlationId, customerId, payment)
    if createPaymentErr != nil {
        errorCode := wuuid.NewUUID()
        h.logger.Error(
            "payment creation failed",
            logattr.Error(createPaymentErr.Error()),
            logattr.ErrorCode(errorCode),
            logattr.CorrelationId(correlationId),
        )
        switch createPaymentErr.Code() {
        case werrors.ResourceAlreadyExistErrorCode:
            return &api.PostPaymentConflict{
                ErrorMessage: "the payment you are trying to create already exist",
                ErrorCode:    errorCode,
            }, nil
        case werrors.ValidationErrorCode:
            return &api.PostPaymentBadRequest{
                ErrorMessage: createPaymentErr.Message(),
                ErrorCode:    errorCode,
            }, nil
        default:
            return &api.PostPaymentInternalServerError{
                ErrorMessage: "unexpected internal error",
                ErrorCode:    errorCode,
            }, nil
        }
    }

    h.logger.Info("payment created",
        logattr.CorrelationId(correlationId),
        logattr.PaymentId(paymentCreated.Data.ID.String()),
    )

    return &paymentCreated.Data, nil
}
