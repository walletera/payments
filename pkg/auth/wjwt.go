package auth

import (
    "crypto/rsa"
    "encoding/json"

    "github.com/golang-jwt/jwt"
)

// WJWT struct represents parsed jwt information.
type WJWT struct {
    UID      string      `json:"uid"`
    State    string      `json:"state"`
    Email    string      `json:"email"`
    Role     string      `json:"role"`
    Level    json.Number `json:"level"`
    Audience []string    `json:"aud,omitempty"`

    jwt.StandardClaims
}

// ParseAndValidate parses token and validates it's jwt signature with given key.
func ParseAndValidate(token string, key *rsa.PublicKey) (WJWT, error) {
    wjwt := WJWT{}

    _, err := jwt.ParseWithClaims(token, &wjwt, func(t *jwt.Token) (interface{}, error) {
        return key, nil
    })

    return wjwt, err
}
