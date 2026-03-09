package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rest-api/configs"
	"rest-api/utils"

	"github.com/redis/go-redis/v9"
)

const (
	AuditQueueKey = "audit:events"
)

type AuditEvent struct {
	UserID     *uint  `json:"user_id"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	Status     string `json:"status"`
	RequestID  string `json:"request_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	DurationMs int64  `json:"duration_ms"`
	Metadata   string `json:"metadata"`
	Timestamp  int64  `json:"timestamp"`
}

type AuditPublisher struct {
	client *redis.Client
}

var auditPublisher *AuditPublisher

func InitAuditPublisher() error {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", configs.AppEnv.RedisHost, configs.AppEnv.RedisPort),
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		utils.LogErrorPlain("Failed to connect to Redis for audit publisher: " + err.Error())
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	auditPublisher = &AuditPublisher{client: client}
	utils.LogInfo("Audit publisher connected to Redis")
	return nil
}

func GetAuditPublisher() *AuditPublisher {
	return auditPublisher
}

func (p *AuditPublisher) Publish(event AuditEvent) error {
	if p == nil || p.client == nil {
		utils.LogErrorPlain("Audit publisher not initialized")
		return fmt.Errorf("audit publisher not initialized")
	}

	event.Timestamp = time.Now().UnixMilli()

	data, err := json.Marshal(event)
	if err != nil {
		utils.LogErrorPlain("Failed to marshal audit event: " + err.Error())
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := p.client.LPush(ctx, AuditQueueKey, data).Err(); err != nil {
		utils.LogErrorPlain("Failed to push audit event to Redis: " + err.Error())
		return fmt.Errorf("failed to push audit event to Redis: %w", err)
	}

	utils.LogInfo(fmt.Sprintf("Audit event published: %s %s - %s", event.Method, event.Path, event.Status))
	return nil
}

func (p *AuditPublisher) Close() error {
	if p == nil || p.client == nil {
		return nil
	}
	return p.client.Close()
}
