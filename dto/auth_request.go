package dto

import "github.com/golang-jwt/jwt/v5"

type RegisterInput struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID  uint   `json:"user_id"`
	TokenID string `json:"token_id"`
}