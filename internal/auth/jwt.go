package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	secret string
	aud    string
	iss    string
}

func NewJwtConfig(secret, aud, iss string) *JWTConfig {
	return &JWTConfig{
		secret: secret,
		aud:    aud,
		iss:    iss,
	}
}

func (j *JWTConfig) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWTConfig) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	},
		jwt.WithExpirationRequired(),                                // Ensure the token is not expired
		jwt.WithAudience(j.aud),                                     // Ensure the token has the correct audience
		jwt.WithIssuer(j.iss),                                       // Ensure the token has the correct issuer
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), // Ensure the token has the correct signing method
	)
}
