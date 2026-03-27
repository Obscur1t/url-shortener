package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID int `json:"uid"`
	jwt.RegisteredClaims
}

func NewToken(key string, userID int, duration time.Duration) (string, error) {
	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(key))
}

func ParseToken(token string, secret string) (int, error) {
	var claims UserClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrMethod
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, jwt.ErrTokenExpired
		}
		return 0, ErrParseToken
	}
	if !t.Valid {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}
