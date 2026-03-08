package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"rest-api/dto"
	"rest-api/services"
	"rest-api/utils"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var input dto.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid input: "+err.Error())
		return
	}

	user, err := services.Register(input.Name, input.Email, input.Password)
	if err != nil {
		if err == services.ErrEmailExists {
			utils.Error(c, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "REGISTRATION_FAILED", "Registration failed")
		return
	}

	utils.Success(c, "User registered successfully", gin.H{"user": user})
}

func Login(c *gin.Context) {
	var input dto.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid input: "+err.Error())
		return
	}

	tokens, err := services.Login(input.Email, input.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			limiter := services.GetLoginLimiter()
			remaining := limiter.GetRemainingAttempts(input.Email)
			if remaining > 0 {
				utils.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", 
					fmt.Sprintf("Invalid email or password. %d attempts remaining", remaining))
			} else {
				utils.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", 
					"Invalid email or password")
			}
			return
		}
		if errors.Is(err, services.ErrAccountLocked) || 
		   strings.Contains(err.Error(), "account temporarily locked") {
			utils.Error(c, http.StatusTooManyRequests, "ACCOUNT_LOCKED", err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, "LOGIN_FAILED", "Login failed")
		return
	}

	utils.SetAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)
	utils.Success(c, "Logged in successfully", gin.H{
		"user": gin.H{
			"id":    tokens.UserID,
			"email": input.Email,
		},
	})
}

func RefreshToken(c *gin.Context) {
	refreshToken, err := utils.GetRefreshToken(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "MISSING_TOKEN", "Refresh token required")
		return
	}

	tokens, err := services.Refresh(refreshToken)
	if err != nil {
		utils.ClearAuthCookies(c)
		if err == services.ErrInvalidToken {
			utils.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid refresh token")
			return
		}
		if err == services.ErrTokenExpired {
			utils.Error(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Refresh token expired")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "REFRESH_FAILED", "Token refresh failed")
		return
	}

	utils.SetAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)
	utils.Success(c, "Token refreshed successfully", nil)
}

func Logout(c *gin.Context) {
	accessToken, err := utils.GetAccessToken(c)
	if err == nil {
		expiry, _ := utils.GetTokenExpiry(accessToken)
		if expiry.IsZero() {
			expiry = time.Now().Add(15 * time.Minute)
		}
		services.Logout(accessToken, expiry)
	}

	utils.ClearAuthCookies(c)
	utils.Success(c, "Logged out successfully", nil)
}

func GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	user, err := services.GetUserByID(userID.(uint))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	utils.Success(c, "User profile retrieved successfully", gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}
