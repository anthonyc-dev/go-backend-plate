package middleware

import (
	"fmt"
	"strings"
	"time"

	"rest-api/services"
	"rest-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
)

// AuditMiddleware logs all requests and publishes audit events to Redis/Postgres
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Set or generate a request ID
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header(RequestIDHeader, requestID)

		// Proceed with request
		c.Next()

		// Measure duration
		durationMs := time.Since(startTime).Milliseconds()

		// Safely extract user ID from context
		var userIDPtr *uint
		if userID, exists := c.Get("user_id"); exists {
			switch v := userID.(type) {
			case uint:
				userIDPtr = new(uint)
				*userIDPtr = v
			case int:
				userIDPtr = new(uint)
				*userIDPtr = uint(v)
			case int64:
				userIDPtr = new(uint)
				*userIDPtr = uint(v)
			case float64:
				userIDPtr = new(uint)
				*userIDPtr = uint(v)
			case *uint:
				userIDPtr = v
			default:
				utils.LogErrorPlain(fmt.Sprintf("Unexpected type for user_id: %T", v))
			}
		}

		// Build audit event
		action := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		resource := extractResource(c.Request.URL.Path)
		status := getStatusCategory(c.Writer.Status())

		event := services.AuditEvent{
			UserID:     userIDPtr,
			Action:     action,
			Resource:   resource,
			IPAddress:  getClientIP(c),
			UserAgent:  c.Request.UserAgent(),
			Status:     status,
			RequestID:  requestID,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			DurationMs: durationMs,
		}

		// Publish event safely
		publisher := services.GetAuditPublisher()
		if publisher != nil {
			if err := publisher.Publish(event); err != nil {
				utils.LogErrorPlain("Failed to publish audit event: " + err.Error())
			}
		}
	}
}

// getClientIP tries headers first, then remote address
func getClientIP(c *gin.Context) string {
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return xri
	}

	if c.Request.RemoteAddr != "" {
		parts := strings.Split(c.Request.RemoteAddr, ":")
		if len(parts) > 1 {
			return strings.Join(parts[:len(parts)-1], ":")
		}
		return c.Request.RemoteAddr
	}

	return "unknown"
}

// extractResource returns the last path segment as the resource
func extractResource(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// getStatusCategory maps HTTP status codes to categories
func getStatusCategory(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "success"
	case status >= 300 && status < 400:
		return "redirect"
	case status >= 400 && status < 500:
		return "client_error"
	case status >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}