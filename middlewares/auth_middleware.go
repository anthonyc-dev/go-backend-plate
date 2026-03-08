package middleware

import (
	"net/http"
	"time"

	"rest-api/database"
	"rest-api/models"
	"rest-api/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := utils.GetAccessToken(c)
		if err != nil {
			utils.LogError(c, http.StatusUnauthorized, "MISSING_TOKEN", "Access token required")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "access token required"})
			c.Abort()
			return
		}

		hashedToken := utils.HashToken(tokenString)

		var blacklistedToken models.TokenBlacklist
		if err := database.DB.Where("token = ?", hashedToken).First(&blacklistedToken).Error; err == nil {
			if blacklistedToken.ExpiresAt.After(time.Now()) {
				utils.LogError(c, http.StatusUnauthorized, "TOKEN_BLACKLISTED", "Token has been revoked")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
				c.Abort()
				return
			}
		}

		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			utils.LogError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Token verification failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("token_id", claims.TokenID)

		c.Next()
	}
}
