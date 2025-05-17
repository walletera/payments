package private

import (
    "context"
    "log/slog"

    privapi "github.com/walletera/payments-types/privateapi"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
    "github.com/walletera/werrors"
)

type Handler struct {
    service *payment.Service
    logger  *slog.Logger
}

var _ privapi.Handler = (*Handler)(nil)

func NewHandler(service *payment.Service, logger *slog.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

func (h *Handler) GetPayment(ctx context.Context, params privapi.GetPaymentParams) (privapi.GetPaymentRes, error) {
    retrievedPayment, getPaymentErr := h.service.GetPayment(ctx, params.PaymentId)
    if getPaymentErr != nil {
        h.logger.Error("failed getting payment",
            logattr.Error(getPaymentErr.Error()),
            logattr.ErrorCode(getPaymentErr.Code()),
            logattr.PaymentId(params.PaymentId.String()),
        )
        switch getPaymentErr.Code() {
        case werrors.ResourceNotFoundErrorCode:
            return &privapi.GetPaymentNotFound{}, nil
        default:
            return &privapi.GetPaymentInternalServerError{}, nil
        }
    }
    return &retrievedPayment, nil
}

func (h *Handler) PatchPayment(ctx context.Context, req *privapi.PaymentUpdate, params privapi.PatchPaymentParams) (privapi.PatchPaymentRes, error) {
    var correlationId string
    if params.XWalleteraCorrelationID.Set {
        correlationId = params.XWalleteraCorrelationID.Value.String()
    } else {
        correlationId = wuuid.NewUUID().String()
    }
    err := h.service.UpdatePayment(ctx, correlationId, *req)
    if err != nil {
        h.logger.Error(
            "payment creation failed",
            logattr.CorrelationId(correlationId),
            logattr.Error(err.Error()),
            logattr.ErrorCode(err.Code()),
        )
        switch err.Code() {
        case werrors.ValidationErrorCode:
            return &privapi.PatchPaymentBadRequest{
                ErrorMessage: err.Message(),
                ErrorCode:    err.Code().String(),
            }, nil
        default:
            return &privapi.PatchPaymentInternalServerError{
                ErrorMessage: "unexpected internal error",
                ErrorCode:    err.Code().String(),
            }, nil
        }
    }
    return &privapi.PatchPaymentOK{}, nil
}

func (h *Handler) PostPayment(ctx context.Context, req *privapi.PostPaymentReq, params privapi.PostPaymentParams) (privapi.PostPaymentRes, error) {
    var correlationId string
    if params.XWalleteraCorrelationID.Set {
        correlationId = params.XWalleteraCorrelationID.Value.String()
    } else {
        correlationId = wuuid.NewUUID().String()
    }
    return h.createPayment(ctx, correlationId, req)
}

func (h *Handler) createPayment(ctx context.Context, correlationId string, paymentCreationRequest *privapi.PostPaymentReq, ) (privapi.PostPaymentRes, error) {
    privPayment := buildPrivPaymentFromPaymentCreationRequest(paymentCreationRequest)
    paymentCreated, createPaymentErr := h.service.CreatePayment(ctx, correlationId, privPayment)
    if createPaymentErr != nil {
        h.logger.Error(
            "payment creation failed",
            logattr.Error(createPaymentErr.Error()),
            logattr.ErrorCode(createPaymentErr.Code()),
            logattr.CorrelationId(correlationId),
        )
        switch createPaymentErr.Code() {
        case werrors.ResourceAlreadyExistErrorCode:
            return &privapi.PostPaymentConflict{
                ErrorMessage: "the payment you are trying to create already exist",
                ErrorCode:    createPaymentErr.Code().String(),
            }, nil
        case werrors.ValidationErrorCode:
            return &privapi.PostPaymentBadRequest{
                ErrorMessage: createPaymentErr.Message(),
                ErrorCode:    createPaymentErr.Code().String(),
            }, nil
        default:
            return &privapi.PostPaymentInternalServerError{
                ErrorMessage: "unexpected internal error",
                ErrorCode:    createPaymentErr.Code().String(),
            }, nil
        }
    }

    h.logger.Info("payment created",
        logattr.CorrelationId(correlationId),
        logattr.PaymentId(paymentCreated.Data.ID.String()),
    )

    return &paymentCreated.Data, nil
}

func buildPrivPaymentFromPaymentCreationRequest(paymentCreationRequest *privapi.PostPaymentReq) privapi.Payment {
    return privapi.Payment{
        ID:          paymentCreationRequest.ID,
        Amount:      paymentCreationRequest.Amount,
        Currency:    paymentCreationRequest.Currency,
        Gateway:     paymentCreationRequest.Gateway,
        Debtor:      paymentCreationRequest.Debtor,
        Beneficiary: paymentCreationRequest.Beneficiary,
        Direction:   paymentCreationRequest.Direction,
        CustomerId:  paymentCreationRequest.CustomerId,
        Status:      paymentCreationRequest.Status,
        ExternalId:  paymentCreationRequest.ExternalId,
        SchemeId:    paymentCreationRequest.SchemeId,
    }
}
