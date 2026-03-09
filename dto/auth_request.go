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

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPInput struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type ResetPasswordInput struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
