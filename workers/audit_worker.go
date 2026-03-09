package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"rest-api/configs"
	"rest-api/database"
	"rest-api/models"
	"rest-api/services"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	AuditQueueKey = "audit:events"
	WorkerName    = "audit-worker"
)

type AuditWorker struct {
	client    *redis.Client
	db        *gorm.DB
	workerID  string
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	processed int64
	failed    int64
}

var auditWorker *AuditWorker

func NewAuditWorker() *AuditWorker {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", configs.AppEnv.RedisHost, configs.AppEnv.RedisPort),
		Password:     "",
		DB:           0,
		PoolSize:     5,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	workerID := fmt.Sprintf("%s-%d", WorkerName, time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())

	return &AuditWorker{
		client:   client,
		db:       database.DB,
		workerID: workerID,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func InitAuditWorker() error {
	worker := NewAuditWorker()

	if err := worker.client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis for audit worker: %w", err)
	}

	auditWorker = worker
	log.Printf("[AUDIT WORKER] Initialized with ID: %s", worker.workerID)
	return nil
}

func GetAuditWorker() *AuditWorker {
	return auditWorker
}

func (w *AuditWorker) Start() {
	if w.isRunning {
		log.Printf("[AUDIT WORKER] Worker already running")
		return
	}

	w.isRunning = true
	w.wg.Add(1)
	go w.run()

	log.Printf("[AUDIT WORKER] Started")
}

func (w *AuditWorker) Stop() {
	if !w.isRunning {
		return
	}

	log.Printf("[AUDIT WORKER] Shutting down gracefully...")
	w.cancel()
	w.wg.Wait()
	w.isRunning = false
	log.Printf("[AUDIT WORKER] Stopped. Processed: %d, Failed: %d", w.processed, w.failed)
}

func (w *AuditWorker) run() {
	defer w.wg.Done()

	log.Printf("[AUDIT WORKER %s] Started processing audit events", w.workerID)

	for {
		select {
		case <-w.ctx.Done():
			w.drainQueue()
			log.Printf("[AUDIT WORKER] Context cancelled, exiting")
			return
		default:
			w.processNextEvent()
		}
	}
}

func (w *AuditWorker) processNextEvent() {
	result, err := w.client.BRPop(w.ctx, 5*time.Second, AuditQueueKey).Result()

	if err != nil {
		if w.ctx.Err() != nil {
			return
		}
		return
	}

	if len(result) < 2 {
		return
	}

	eventData := result[1]

	var event services.AuditEvent
	if err := json.Unmarshal([]byte(eventData), &event); err != nil {
		log.Printf("[AUDIT WORKER] Failed to unmarshal event: %v", err)
		w.failed++
		return
	}

	if err := w.saveToDatabase(event); err != nil {
		log.Printf("[AUDIT WORKER] Failed to save event to database: %v", err)
		w.failed++

		w.requeueEvent(eventData)
		return
	}

	w.processed++
}

func (w *AuditWorker) saveToDatabase(event services.AuditEvent) error {
	var metadata models.JSONMap
	if event.Metadata != "" {
		if err := json.Unmarshal([]byte(event.Metadata), &metadata); err != nil {
			metadata = models.JSONMap{}
		}
	}

	auditLog := models.AuditLog{
		UserID:     event.UserID,
		Action:     event.Action,
		Resource:   event.Resource,
		IPAddress:  event.IPAddress,
		UserAgent:  event.UserAgent,
		Status:     event.Status,
		RequestID:  event.RequestID,
		Method:     event.Method,
		Path:       event.Path,
		DurationMs: event.DurationMs,
		Metadata:   metadata,
		CreatedAt:  time.UnixMilli(event.Timestamp),
	}

	tx := w.db.Create(&auditLog)
	if tx.Error != nil {
		return fmt.Errorf("failed to insert audit log: %w", tx.Error)
	}

	return nil
}

func (w *AuditWorker) requeueEvent(eventData string) {
	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := w.client.LPush(ctx, AuditQueueKey, eventData).Err(); err != nil {
		log.Printf("[AUDIT WORKER] Failed to requeue event: %v", err)
	}
}

func (w *AuditWorker) drainQueue() {
	log.Printf("[AUDIT WORKER] Draining remaining queue items...")

	for {
		result, err := w.client.RPopLPush(w.ctx, AuditQueueKey, AuditQueueKey+":drain").Result()
		if err != nil {
			break
		}

		var event services.AuditEvent
		if err := json.Unmarshal([]byte(result), &event); err != nil {
			continue
		}

		if err := w.saveToDatabase(event); err != nil {
			log.Printf("[AUDIT WORKER] Failed to save drained event: %v", err)
		}
	}
}

func (w *AuditWorker) GetStats() map[string]int64 {
	return map[string]int64{
		"processed": w.processed,
		"failed":    w.failed,
		"running":   boolToInt64(w.isRunning),
	}
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func (w *AuditWorker) Close() error {
	w.Stop()
	if w.client != nil {
		return w.client.Close()
	}
	return nil
}

func GetAuditLogs(filters map[string]interface{}, page, limit int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := database.DB.Model(&models.AuditLog{})

	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}

	if action, ok := filters["action"]; ok {
		query = query.Where("action LIKE ?", "%"+action.(string)+"%")
	}

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}

	if resource, ok := filters["resource"]; ok {
		query = query.Where("resource = ?", resource)
	}

	if startDate, ok := filters["start_date"]; ok {
		query = query.Where("created_at >= ?", startDate.(time.Time))
	}

	if endDate, ok := filters["end_date"]; ok {
		query = query.Where("created_at <= ?", endDate.(time.Time))
	}

	if ipAddress, ok := filters["ip_address"]; ok {
		query = query.Where("ip_address LIKE ?", "%"+ipAddress.(string)+"%")
	}

	query.Count(&total)

	offset := (page - 1) * limit
	query = query.Order("created_at DESC").Offset(offset).Limit(limit)

	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func SearchAuditLogs(query string, page, limit int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	searchQuery := "%" + strings.TrimSpace(query) + "%"

	db := database.DB.Model(&models.AuditLog{}).
		Where("action ILIKE ? OR resource ILIKE ? OR ip_address ILIKE ? OR user_agent ILIKE ? OR request_id ILIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)

	db.Count(&total)

	offset := (page - 1) * limit
	db = db.Order("created_at DESC").Offset(offset).Limit(limit)

	if err := db.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
