package main

import (
	"net/http"
	"os"
	"os/signal"
	"rest-api/configs"
	"rest-api/controllers"
	"rest-api/database"
	"rest-api/routes"
	"rest-api/services"
	"rest-api/workers"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	configs.LoadEnv()

	database.ConnectDB()
	controllers.InitDB(database.DB)

	if err := services.InitAuditPublisher(); err != nil {
		println("Warning: Failed to initialize audit publisher:", err.Error())
	}

	if err := workers.InitAuditWorker(); err != nil {
		println("Warning: Failed to initialize audit worker:", err.Error())
	} else {
		workers.GetAuditWorker().Start()
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "CORS working"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "go-api",
			"time":    time.Now(),
		})
	})

	routes.AuthRoutes(r)
	routes.UserRoutes(r)

	port := configs.AppEnv.Port
	if port == "" {
		port = "8080"
	}

	go func() {
		if err := r.Run(":" + port); err != nil {
			println("Server error:", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	println("\nShutting down...")

	if auditWorker := workers.GetAuditWorker(); auditWorker != nil {
		auditWorker.Stop()
	}

	println("Server stopped")
}
