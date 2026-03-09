package services

import (
	"errors"
	"sync"
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
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrTokenExpired       = errors.New("refresh token expired")
	ErrOTPExpired         = errors.New("OTP has expired")
	ErrInvalidOTP         = errors.New("invalid OTP")
	ErrOTPAlreadyUsed     = errors.New("OTP already used")
	ErrTooManyAttempts    = errors.New("too many OTP attempts")
	ErrUserNotFound       = errors.New("user not found")
)

type OTPRateLimiter struct {
	attempts map[string]int
	lock     sync.RWMutex
	limit    int
	window   time.Duration
}

func NewOTPRateLimiter(limit int, window time.Duration) *OTPRateLimiter {
	return &OTPRateLimiter{
		attempts: make(map[string]int),
		limit:    limit,
		window:   window,
	}
}

func (r *OTPRateLimiter) Check(email string) error {
	r.lock.RLock()
	attempts := r.attempts[email]
	r.lock.RUnlock()

	if attempts >= r.limit {
		return ErrTooManyAttempts
	}
	return nil
}

func (r *OTPRateLimiter) Record(email string) {
	r.lock.Lock()
	r.attempts[email]++
	r.lock.Unlock()

	time.AfterFunc(r.window, func() {
		r.lock.Lock()
		delete(r.attempts, email)
		r.lock.Unlock()
	})
}

var otpLimiter = NewOTPRateLimiter(3, 5*time.Minute)

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

func RequestForgotPassword(email string) error {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if err := otpLimiter.Check(email); err != nil {
		utils.LogErrorPlain("OTP rate limit exceeded for email: " + email)
		return err
	}

	existingOTP := models.PasswordResetOTP{}
	database.DB.Where("email = ? AND is_used = ? AND expires_at > ?", email, false, time.Now()).
		First(&existingOTP)
	if existingOTP.ID != 0 {
		utils.LogErrorPlain("Active OTP already exists for email: " + email)
		return errors.New("OTP already sent, please wait 5 minutes")
	}

	otp, err := GenerateSecureOTP(6)
	if err != nil {
		return err
	}

	hashedOTP := HashOTP(otp)
	expiresAt := GetOTPExpiry()

	otpRecord := models.PasswordResetOTP{
		Email:     email,
		OTP:       hashedOTP,
		ExpiresAt: expiresAt,
		Attempts:  0,
		IsUsed:    false,
	}

	if err := database.DB.Create(&otpRecord).Error; err != nil {
		return err
	}

	emailService := NewEmailService()
	if err := emailService.SendOTP(email, otp); err != nil {
		utils.LogErrorPlain("Failed to send OTP email: " + err.Error())
		return errors.New("failed to send OTP email")
	}

	otpLimiter.Record(email)

	utils.LogInfo("OTP sent successfully to email: " + email)
	return nil
}

func VerifyOTP(email, otp string) error {
	var otpRecord models.PasswordResetOTP
	if err := database.DB.Where("email = ? AND is_used = ?", email, false).
		Order("created_at DESC").First(&otpRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.LogErrorPlain("No active OTP found for email: " + email)
			return ErrInvalidOTP
		}
		return err
	}

	if time.Now().After(otpRecord.ExpiresAt) {
		utils.LogErrorPlain("OTP expired for email: " + email)
		return ErrOTPExpired
	}

	hashedOTP := HashOTP(otp)
	if hashedOTP != otpRecord.OTP {
		otpRecord.Attempts++
		database.DB.Save(&otpRecord)

		if otpRecord.Attempts >= 3 {
			database.DB.Delete(&otpRecord)
			utils.LogErrorPlain("Too many OTP attempts for email: " + email)
			return ErrTooManyAttempts
		}

		utils.LogErrorPlain("Invalid OTP attempt for email: " + email)
		return ErrInvalidOTP
	}

	otpRecord.IsUsed = true
	database.DB.Save(&otpRecord)

	return nil
}

func ResetPassword(email, newPassword string) error {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}

	database.DB.Where("email = ?", email).Delete(&models.PasswordResetOTP{})

	utils.LogInfo("Password reset successfully for email: " + email)
	return nil
}
