package auth

import (
    "crypto/rsa"
    "os"

    "github.com/golang-jwt/jwt"
)

func ReadRSAPrivKeyFromFile(path string) (*rsa.PrivateKey, error) {
    pem, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    key, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
    if err != nil {
        return nil, err
    }

    return key, nil
}
