package public

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/google/uuid"
    _ "github.com/walletera/payments-types/builders/publicapi"
    privconv "github.com/walletera/payments-types/converters/privateapi"
    pubapiconv "github.com/walletera/payments-types/converters/publicapi"
    privapi "github.com/walletera/payments-types/privateapi"
    pubapi "github.com/walletera/payments-types/publicapi"
    "github.com/walletera/payments/internal/domain/payment"
    "github.com/walletera/payments/pkg/logattr"
    "github.com/walletera/payments/pkg/wuuid"
    "github.com/walletera/werrors"
)

type Handler struct {
    service *payment.Service
    logger  *slog.Logger
}

func (h *Handler) ListPayments(ctx context.Context, params pubapi.ListPaymentsParams) (pubapi.ListPaymentsRes, error) {
    return &pubapi.ListPaymentsMethodNotAllowed{}, nil
}

var _ pubapi.Handler = (*Handler)(nil)

func NewHandler(service *payment.Service, logger *slog.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

func (h *Handler) GetPayment(ctx context.Context, params pubapi.GetPaymentParams) (pubapi.GetPaymentRes, error) {
    _, err := getCustomerIdFromCtx(ctx)
    if err != nil {
        h.logger.Warn(err.Error())
        return &pubapi.GetPaymentUnauthorized{}, nil
    }
    retrievedPayment, getPaymentErr := h.service.GetPayment(ctx, params.PaymentId)
    if getPaymentErr != nil {
        h.logger.Error("failed getting payment",
            logattr.Error(getPaymentErr.Error()),
            logattr.ErrorCode(getPaymentErr.Code()),
            logattr.PaymentId(params.PaymentId.String()),
        )
        switch getPaymentErr.Code() {
        case werrors.ResourceNotFoundErrorCode:
            return &pubapi.GetPaymentNotFound{}, nil
        default:
            return &pubapi.GetPaymentInternalServerError{}, nil
        }
    }
    return buildPublicPaymentFromPrivatePayment(retrievedPayment), nil
}

func (h *Handler) PostPayment(ctx context.Context, req *pubapi.PostPaymentReq, params pubapi.PostPaymentParams) (pubapi.PostPaymentRes, error) {
    var correlationId string
    if params.XWalleteraCorrelationID.Set {
        correlationId = params.XWalleteraCorrelationID.Value.String()
    } else {
        correlationId = wuuid.NewUUID().String()
    }
    customerId, err := getCustomerIdFromCtx(ctx)
    if err != nil {
        h.logger.Warn(err.Error())
        return &pubapi.PostPaymentUnauthorized{}, nil
    }
    parsedCustomerId, err := uuid.Parse(customerId)
    if err != nil {
        h.logger.Error(
            "error parsing customerId: "+err.Error(),
            logattr.CorrelationId(correlationId),
        )
        return &pubapi.PostPaymentInternalServerError{
            ErrorMessage: "unexpected internal error",
        }, nil
    }
    return h.CreatePayment(ctx, correlationId, parsedCustomerId, req)
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

func (h *Handler) CreatePayment(ctx context.Context, correlationId string, customerId uuid.UUID, paymentCreationRequest *pubapi.PostPaymentReq, ) (pubapi.PostPaymentRes, error) {
    privPayment, err := buildPrivPaymentFromPaymentCreationRequest(paymentCreationRequest)
    if err != nil {
        return &pubapi.PostPaymentBadRequest{
            ErrorMessage: err.Message(),
            ErrorCode:    err.Code().String(),
        }, nil
    }
    privPayment.Direction = privapi.DirectionOutbound
    privPayment.CustomerId = customerId
    privPayment.Status = privapi.PaymentStatusPending
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
            return &pubapi.PostPaymentConflict{
                ErrorMessage: "the payment you are trying to create already exist",
                ErrorCode:    createPaymentErr.Code().String(),
            }, nil
        case werrors.ValidationErrorCode:
            return &pubapi.PostPaymentBadRequest{
                ErrorMessage: createPaymentErr.Message(),
                ErrorCode:    createPaymentErr.Code().String(),
            }, nil
        default:
            return &pubapi.PostPaymentInternalServerError{
                ErrorMessage: "unexpected internal error",
                ErrorCode:    createPaymentErr.Code().String(),
            }, nil
        }
    }

    h.logger.Info("payment created",
        logattr.CorrelationId(correlationId),
        logattr.PaymentId(paymentCreated.Data.ID.String()),
    )

    return buildPublicPaymentFromPrivatePayment(privPayment), nil
}

func buildPrivPaymentFromPaymentCreationRequest(paymentCreationRequest *pubapi.PostPaymentReq) (privapi.Payment, werrors.WError) {
    return privapi.Payment{
        ID:          paymentCreationRequest.ID,
        Amount:      paymentCreationRequest.Amount,
        Currency:    privapi.Currency(paymentCreationRequest.Currency),
        Gateway:     privapi.Gateway(paymentCreationRequest.Gateway),
        Debtor:      pubapiconv.Convert(paymentCreationRequest.Debtor),
        Beneficiary: pubapiconv.Convert(paymentCreationRequest.Beneficiary),
    }, nil
}

func buildPublicPaymentFromPrivatePayment(p privapi.Payment) *pubapi.Payment {
    return &pubapi.Payment{
        ID:          p.ID,
        Amount:      p.Amount,
        Currency:    pubapi.Currency(p.Currency),
        Debtor:      privconv.Convert(p.Debtor),
        Beneficiary: privconv.Convert(p.Beneficiary),
        Direction:   pubapi.Direction(p.Direction),
        Status:      pubapi.PaymentStatus(p.Status),
        Gateway:     pubapi.Gateway(p.Gateway),
        CreatedAt:   p.CreatedAt,
        UpdatedAt:   p.UpdatedAt,
    }
}
