package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	env "rest-api/configs"
	"rest-api/dto"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateAccessToken(userID uint) (string, string, error) {
	tokenID := uuid.New().String()

	claims := dto.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        tokenID,
		},
		UserID:  userID,
		TokenID: tokenID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(env.AppEnv.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return tokenString, tokenID, nil
}

func GenerateSecureRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func VerifyToken(tokenString string) (*dto.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &dto.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(env.AppEnv.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*dto.TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := VerifyToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}
