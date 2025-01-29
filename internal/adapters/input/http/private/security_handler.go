package private

import (
    "context"

    "github.com/walletera/payments-types/api"
)

type SecurityHandler struct {
}

func NewSecurityHandler() *SecurityHandler {
    return &SecurityHandler{}
}

func (s *SecurityHandler) HandleBearerAuth(ctx context.Context, operationName api.OperationName, t api.BearerAuth) (context.Context, error) {
    // TODO implement internal security handler
    return ctx, nil
}
