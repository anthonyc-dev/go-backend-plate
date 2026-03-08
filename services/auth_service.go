package services

import (
	"errors"
	"time"

	"rest-api/database"
	"rest-api/models"
	"rest-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessID     string
	RefreshID    string
	UserID       uint
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already registered")
	ErrInvalidToken      = errors.New("invalid refresh token")
	ErrTokenExpired      = errors.New("refresh token expired")
)

func Register(name, email, password string) (*models.User, error) {

	var existing models.User
	if err := database.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func Login(email, password string) (*TokenPair, error) {
	limiter := GetLoginLimiter()

	if err := limiter.CheckBlocked(email); err != nil {
		return nil, err
	}

	var user models.User

	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := utils.CheckPassword(password, user.Password); err != nil {
		limiter.RecordFailedAttempt(email)
		return nil, ErrInvalidCredentials
	}

	limiter.ResetAttempts(email)

	accessToken, accessID, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateSecureRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshID := uuid.New().String()
	hashedRefreshToken := utils.HashToken(refreshToken)

	expiry := time.Now().Add(7 * 24 * time.Hour)
	if err := database.DB.Create(&models.RefreshToken{
		UserID:    user.ID,
		Token:     hashedRefreshToken,
		TokenID:   refreshID,
		ExpiresAt: expiry,
	}).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessID:     accessID,
		RefreshID:    refreshID,
		UserID:       user.ID,
	}, nil
}

func Refresh(oldRefreshToken string) (*TokenPair, error) {

	hashedToken := utils.HashToken(oldRefreshToken)

	var storedToken models.RefreshToken

	if err := database.DB.Where("token = ?", hashedToken).First(&storedToken).Error; err != nil {
		return nil, ErrInvalidToken
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		database.DB.Delete(&storedToken)
		return nil, ErrTokenExpired
	}

	newAccess, accessID, err := utils.GenerateAccessToken(storedToken.UserID)
	if err != nil {
		return nil, err
	}

	newRefresh, err := utils.GenerateSecureRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshID := uuid.New().String()

	database.DB.Delete(&storedToken)

	hashedNewRefresh := utils.HashToken(newRefresh)

	expiry := time.Now().Add(7 * 24 * time.Hour)
	if err := database.DB.Create(&models.RefreshToken{
		UserID:    storedToken.UserID,
		Token:     hashedNewRefresh,
		TokenID:   refreshID,
		ExpiresAt: expiry,
	}).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		AccessID:     accessID,
		RefreshID:    refreshID,
		UserID:       storedToken.UserID,
	}, nil
}

func Logout(token string, expiry time.Time) error {

	hashedToken := utils.HashToken(token)

	var blacklisted models.TokenBlacklist

	if err := database.DB.Where("token = ?", hashedToken).First(&blacklisted).Error; err == nil {
		return nil
	}

	return database.DB.Create(&models.TokenBlacklist{
		Token:     hashedToken,
		ExpiresAt: expiry,
	}).Error
}

func GetUserByID(userID uint) (*models.User, error) {

	var user models.User

	if err := database.DB.First(&user, userID).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return &user, nil
}
