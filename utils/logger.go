package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
)

type LogLevel string

const (
	INFO    LogLevel = "INFO"
	SUCCESS LogLevel = "SUCCESS"
	WARN    LogLevel = "WARN"
	ERROR   LogLevel = "ERROR"
	DEBUG   LogLevel = "DEBUG"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Status    int    `json:"status"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message"`
	Path      string `json:"path"`
	Method    string `json:"method"`
}

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	infoLogger = log.New(os.Stdout, "", 0)
	errorLogger = log.New(os.Stderr, "", 0)
}

func getLevelColor(level LogLevel) string {
	switch level {
	case INFO, DEBUG:
		return Blue
	case SUCCESS:
		return Green
	case WARN:
		return Yellow
	case ERROR:
		return Red
	default:
		return White
	}
}

func formatLog(entry LogEntry) string {
	color := getLevelColor(LogLevel(entry.Level))
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf("%s[%s]%s %s[%s]%s | %s[%d]%s | %s%-6s%s | %s%s%s | %s%s%s",
		White, timestamp, Reset,
		color, entry.Level, Reset,
		Bold, entry.Status, Reset,
		Magenta, entry.Method, Reset,
		White, entry.Path, Reset,
		color, entry.Message, Reset,
	)
}

func formatLogPlain(entry LogEntry) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf("[%s] %s | %d | %-6s | %s | %s",
		timestamp,
		entry.Level,
		entry.Status,
		entry.Method,
		entry.Path,
		entry.Message,
	)
}

func Log(level LogLevel, status int, errorCode, message, path, method string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     string(level),
		Status:    status,
		ErrorCode: errorCode,
		Message:   message,
		Path:      path,
		Method:    method,
	}

	if level == ERROR {
		errorLogger.Println(formatLogPlain(entry))
	} else {
		infoLogger.Println(formatLogPlain(entry))
	}
}

func LogRequest(c *gin.Context, status int, errorCode, message string) {
	Log(
		ERROR,
		status,
		errorCode,
		message,
		c.Request.URL.Path,
		c.Request.Method,
	)
}

func LogSuccess(c *gin.Context, message string) {
	Log(
		SUCCESS,
		http.StatusOK,
		"",
		message,
		c.Request.URL.Path,
		c.Request.Method,
	)
}

func LogError(c *gin.Context, status int, errorCode, message string) {
	Log(
		ERROR,
		status,
		errorCode,
		message,
		c.Request.URL.Path,
		c.Request.Method,
	)
}

func LogInfo(message string) {
	Log(INFO, 0, "", message, "-", "-")
}

func LogWarn(message string) {
	Log(WARN, 0, "", message, "-", "-")
}

func LogErrorPlain(message string) {
	Log(ERROR, 0, "", message, "-", "-")
}
