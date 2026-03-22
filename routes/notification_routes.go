package routes

import (
	"rest-api/controllers"
	middleware "rest-api/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

func NotificationRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(middleware.AuditMiddleware())

	notifications := api.Group("/notifications")
	notifications.Use(middleware.RateLimiterMiddleware(100, time.Minute))
	{
		notifications.POST("/push/register", middleware.AuthMiddleware(), controllers.RegisterPushToken)
		notifications.POST("/push/unregister", middleware.AuthMiddleware(), controllers.DeletePushToken)

		notifications.POST("/notify", middleware.AuthMiddleware(), controllers.NotifyUser)
		notifications.POST("/broadcast", middleware.AuthMiddleware(), controllers.Broadcast)

		notifications.GET("/", middleware.AuthMiddleware(), controllers.GetMyNotifications)
		notifications.GET("/:id", middleware.AuthMiddleware(), controllers.GetNotificationByID)
		notifications.GET("/unread-count", middleware.AuthMiddleware(), controllers.GetUnreadCount)

		notifications.PUT("/:id/read", middleware.AuthMiddleware(), controllers.MarkAsRead)
		notifications.PUT("/read-all", middleware.AuthMiddleware(), controllers.MarkAllAsRead)
	}
}
