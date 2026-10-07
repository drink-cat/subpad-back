package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SessionClaims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Sign(secret string, userID int64, username string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt secret is empty")
	}
	if ttl <= 0 {
		ttl = 72 * time.Hour
	}
	now := time.Now()
	claims := SessionClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

func Parse(secret, raw string) (int64, string, error) {
	if secret == "" {
		return 0, "", fmt.Errorf("jwt secret is empty")
	}
	claims := &SessionClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, "", fmt.Errorf("parse jwt: %w", err)
	}
	if token == nil || !token.Valid || claims.UserID <= 0 {
		return 0, "", fmt.Errorf("parse jwt: invalid token")
	}
	return claims.UserID, claims.Username, nil
}
