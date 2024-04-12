package tests

import (
    "context"
    "net/http/httptest"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "github.com/walletera/payments/api"
)

func TestClient_PatchWithdrawal(t *testing.T) {
    handlerMock := NewMockHandler(t)

    withdrawalId, err := uuid.NewUUID()
    require.NoError(t, err)

    externalId, err := uuid.NewUUID()
    require.NoError(t, err)

    withdrawalPatchBody := &api.WithdrawalPatchBody{
        ExternalID: api.OptUUID{
            Value: externalId,
            Set:   true,
        },
        Status: api.WithdrawalPatchBodyStatusConfirmed,
    }

    patchWithdrawalParams := api.PatchWithdrawalParams{
        WithdrawalId: withdrawalId,
    }

    handlerMock.EXPECT().
        PatchWithdrawal(mock.Anything, withdrawalPatchBody, patchWithdrawalParams).
        Return(&api.PatchWithdrawalOK{}, nil)

    paymentsServer, err := api.NewServer(handlerMock)
    require.NoError(t, err)

    ts := httptest.NewServer(paymentsServer)
    defer ts.Close()

    paymentsClient, err := api.NewClient(ts.URL)
    require.NoError(t, err)

    _, err = paymentsClient.PatchWithdrawal(context.Background(), withdrawalPatchBody, patchWithdrawalParams)

    require.NoError(t, err)
}
