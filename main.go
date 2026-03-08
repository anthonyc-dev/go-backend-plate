package main

import (
	"net/http"
	"rest-api/configs"
	"rest-api/controllers"
	"rest-api/database"
	"rest-api/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	configs.LoadEnv()

	database.ConnectDB()
	controllers.InitDB(database.DB)

	r := gin.Default()

	//middlewares
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

	r.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "CORS workings"})
	})

		// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "go-api",
			"time":    time.Now(),
		})
	})

	//routes
	routes.AuthRoutes(r)
	routes.UserRoutes(r)

	port := configs.AppEnv.Port
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
