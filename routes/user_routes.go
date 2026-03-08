package routes

import (
	"rest-api/controllers"
	middleware "rest-api/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	api := r.Group("/api")

	users := api.Group("/users")
	users.Use(middleware.RateLimiterMiddleware(30, time.Minute))
	{
		users.GET("/", middleware.AuthMiddleware(), controllers.GetUsers)
		users.GET("/:id", middleware.AuthMiddleware(), controllers.GetUser)
		users.POST("/", middleware.AuthMiddleware(), controllers.CreateUser)
		users.PUT("/:id", middleware.AuthMiddleware(), controllers.UpdateUser)
		users.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeleteUser)
	}
}
