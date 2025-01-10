package auth

import (
    "os"
    "strings"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestParseAndValidate_ExpiredToken(t *testing.T) {
    token, err := os.ReadFile("testdata/token.txt")
    require.NoError(t, err)

    key, err := ReadRSAPrivKeyFromFile("testdata/test_key.pem")
    require.NoError(t, err)

    wjwt, err := ParseAndValidate(string(token), &key.PublicKey)
    require.Error(t, err)
    require.True(t, strings.Contains(err.Error(), "token is expired"))
    require.Equal(t, "federicoamoya@gmail.com", wjwt.Email)
}

func TestParseAndValidate_DifferentKey(t *testing.T) {
    token, err := os.ReadFile("testdata/token.txt")
    require.NoError(t, err)

    key, err := ReadRSAPrivKeyFromFile("testdata/wrong_key.pem")
    require.NoError(t, err)

    _, err = ParseAndValidate(string(token), &key.PublicKey)
    require.Error(t, err)
    require.True(t, strings.Contains(err.Error(), "verification error"))
}
