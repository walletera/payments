package http

import (
    "context"
    "crypto/rsa"
    "fmt"

    "github.com/walletera/payments-types/api"
    "github.com/walletera/payments/pkg/auth"
)

const WJWTCustomerIdCtxKey = "customer_id"

type SecurityHandler struct {
    pubKey *rsa.PublicKey
}

func NewSecurityHandler(pubKey *rsa.PublicKey) *SecurityHandler {
    return &SecurityHandler{
        pubKey: pubKey,
    }
}

func (s *SecurityHandler) HandleBearerAuth(ctx context.Context, operationName api.OperationName, t api.BearerAuth) (context.Context, error) {
    wjwt, err := auth.ParseAndValidate(t.GetToken(), s.pubKey)
    if err != nil {
        return nil, err
    }
    if len(wjwt.UID) == 0 {
        return nil, fmt.Errorf("uid is missing")
    }
    if wjwt.State != "active" {
        return nil, fmt.Errorf("customer is not active")
    }
    return context.WithValue(ctx, WJWTCustomerIdCtxKey, wjwt.UID), nil
}
