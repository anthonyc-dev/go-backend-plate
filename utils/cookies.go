package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) error {
	expiry := time.Now().Add(7 * 24 * time.Hour)
	maxAge := int(time.Until(expiry).Seconds())

	cookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)

	cookie = &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)

	return nil
}

func GetAccessToken(c *gin.Context) (string, error) {
	token, err := c.Cookie(AccessTokenCookie)
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetRefreshToken(c *gin.Context) (string, error) {
	token, err := c.Cookie(RefreshTokenCookie)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ClearAuthCookies(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)

	cookie = &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)
}

func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func Error(c *gin.Context, status int, errorCode, message string) {
	LogError(c, status, errorCode, message)
	c.JSON(status, gin.H{"error": message})
}

func Success(c *gin.Context, message string, data interface{}) {
	LogSuccess(c, message)
	if data != nil {
		c.JSON(http.StatusOK, data)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message})
	}
}
