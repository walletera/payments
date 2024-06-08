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
        ExternalId: api.OptUUID{
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

func TestClient_PostDeposit(t *testing.T) {
    handlerMock := NewMockHandler(t)

    depositlId, err := uuid.NewUUID()
    require.NoError(t, err)

    customerId, err := uuid.NewUUID()
    require.NoError(t, err)

    externalId, err := uuid.NewUUID()
    require.NoError(t, err)

    depositPostBody := &api.DepositPostBody{
        ID:         depositlId,
        Amount:     100,
        Currency:   "usd",
        CustomerId: customerId,
        ExternalId: externalId,
    }

    handlerMock.EXPECT().
        PostDeposit(mock.Anything, depositPostBody).
        Return(&api.PostDepositCreated{}, nil)

    paymentsServer, err := api.NewServer(handlerMock)
    require.NoError(t, err)

    ts := httptest.NewServer(paymentsServer)
    defer ts.Close()

    paymentsClient, err := api.NewClient(ts.URL)
    require.NoError(t, err)

    _, err = paymentsClient.PostDeposit(context.Background(), depositPostBody)

    require.NoError(t, err)
}
