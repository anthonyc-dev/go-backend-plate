package routes

import (
	"rest-api/controllers"
	middleware "rest-api/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	api := r.Group("/api")

	auth := api.Group("")
	auth.Use(middleware.RateLimiterMiddleware(5, time.Minute))
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/forgot-password", controllers.ForgotPassword)
		auth.POST("/verify-otp", controllers.VerifyOTP)
		auth.POST("/reset-password", controllers.ResetPassword)
	}

	api.POST("/refresh", controllers.RefreshToken)
	api.POST("/logout", controllers.Logout)
	api.GET("/me", middleware.AuthMiddleware(), middleware.RateLimiterMiddleware(5, time.Minute), controllers.GetMe)
}
