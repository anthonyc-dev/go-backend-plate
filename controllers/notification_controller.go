package controllers

import (
	"encoding/json"
	"net/http"
	"rest-api/database"
	"rest-api/models"
	"rest-api/services"
	"rest-api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	DB *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{DB: db}
}

func RegisterPushToken(c *gin.Context) {
	var body struct {
		PushToken string `json:"push_token" validate:"required"`
		Device    string `json:"device" validate:"required,oneof=ios android"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	if err := utils.ValidateStruct(&body); err != nil {
		utils.LogError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	userID := c.GetUint("user_id")

	token := models.PushToken{
		UserID: userID,
		Token:  body.PushToken,
		Device: body.Device,
	}

	if err := database.DB.Where("token = ?", body.PushToken).Assign(token).FirstOrCreate(&token).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "SAVE_FAILED", "Failed to save push token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save push token"})
		return
	}

	utils.LogSuccess(c, "Push token saved")
	c.JSON(http.StatusOK, gin.H{"message": "Push token saved"})
}

func NotifyUser(c *gin.Context) {
	var body struct {
		UserID uint                   `json:"user_id" validate:"required"`
		Title  string                 `json:"title" validate:"required"`
		Body   string                 `json:"body" validate:"required"`
		Type   string                 `json:"type" validate:"omitempty,oneof=system user chat"`
		Data   map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	if err := utils.ValidateStruct(&body); err != nil {
		utils.LogError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	dataJSON := ""
	if body.Data != nil {
		jsonBytes, _ := json.Marshal(body.Data)
		dataJSON = string(jsonBytes)
	}

	notifType := body.Type
	if notifType == "" {
		notifType = "user"
	}

	notif := models.Notification{
		UserID:   &body.UserID,
		IsGlobal: false,
		Title:    body.Title,
		Body:     body.Body,
		Type:     notifType,
		Data:     dataJSON,
		IsRead:   false,
	}

	if err := database.DB.Create(&notif).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create notification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	var tokens []models.PushToken
	if err := database.DB.Where("user_id = ?", body.UserID).Find(&tokens).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "FETCH_TOKENS_FAILED", "Failed to fetch push tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch push tokens"})
		return
	}

	var messages []services.ExpoMessage
	for _, t := range tokens {
		messages = append(messages, services.ExpoMessage{
			To:    t.Token,
			Title: body.Title,
			Body:  body.Body,
			Data:  body.Data,
		})
	}

	if err := services.SendPush(messages); err != nil {
		utils.LogErrorPlain("Push notification failed: " + err.Error())
	}

	utils.LogSuccess(c, "Notification sent")
	c.JSON(http.StatusOK, gin.H{"message": "Notification sent", "notification": notif})
}

func Broadcast(c *gin.Context) {
	var body struct {
		Title string                 `json:"title" validate:"required"`
		Body  string                 `json:"body" validate:"required"`
		Type  string                 `json:"type" validate:"omitempty,oneof=system user chat"`
		Data  map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	if err := utils.ValidateStruct(&body); err != nil {
		utils.LogError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	dataJSON := ""
	if body.Data != nil {
		jsonBytes, _ := json.Marshal(body.Data)
		dataJSON = string(jsonBytes)
	}

	notifType := body.Type
	if notifType == "" {
		notifType = "system"
	}

	notif := models.Notification{
		IsGlobal: true,
		Title:    body.Title,
		Body:     body.Body,
		Type:     notifType,
		Data:     dataJSON,
		IsRead:   false,
	}

	if err := database.DB.Create(&notif).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create notification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	var tokens []models.PushToken
	if err := database.DB.Find(&tokens).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "FETCH_TOKENS_FAILED", "Failed to fetch push tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch push tokens"})
		return
	}

	var messages []services.ExpoMessage
	for _, t := range tokens {
		messages = append(messages, services.ExpoMessage{
			To:    t.Token,
			Title: body.Title,
			Body:  body.Body,
			Data:  body.Data,
		})
	}

	if err := services.SendPush(messages); err != nil {
		utils.LogErrorPlain("Broadcast push notification failed: " + err.Error())
	}

	utils.LogSuccess(c, "Broadcast sent")
	c.JSON(http.StatusOK, gin.H{"message": "Broadcast sent", "notification": notif})
}

func GetMyNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	var notifications []models.Notification
	if err := database.DB.
		Where("user_id = ? OR is_global = ?", userID, true).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch notifications")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	utils.LogSuccess(c, "Notifications retrieved")
	c.JSON(http.StatusOK, notifications)
}

func GetNotificationByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")

	var notification models.Notification
	if err := database.DB.
		Where("id = ? AND (user_id = ? OR is_global = ?)", id, userID, true).
		First(&notification).Error; err != nil {
		utils.LogError(c, http.StatusNotFound, "NOT_FOUND", "Notification not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	utils.LogSuccess(c, "Notification retrieved")
	c.JSON(http.StatusOK, notification)
}

func MarkAsRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")

	result := database.DB.Model(&models.Notification{}).
		Where("id = ? AND (user_id = ? OR is_global = ?)", id, userID, true).
		Update("is_read", true)

	if result.RowsAffected == 0 {
		utils.LogError(c, http.StatusNotFound, "NOT_FOUND", "Notification not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	utils.LogSuccess(c, "Notification marked as read")
	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func MarkAllAsRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? OR is_global = ?", userID, true).
		Update("is_read", true).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to mark all as read")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	utils.LogSuccess(c, "All notifications marked as read")
	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

func DeletePushToken(c *gin.Context) {
	var body struct {
		PushToken string `json:"push_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	userID := c.GetUint("user_id")

	if err := database.DB.Where("token = ? AND user_id = ?", body.PushToken, userID).
		Delete(&models.PushToken{}).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete push token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete push token"})
		return
	}

	utils.LogSuccess(c, "Push token deleted")
	c.JSON(http.StatusOK, gin.H{"message": "Push token deleted"})
}

func GetUnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")

	var count int64
	if err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? OR is_global = ?", userID, true).
		Where("is_read = ?", false).
		Count(&count).Error; err != nil {
		utils.LogError(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count notifications")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count notifications"})
		return
	}

	utils.LogSuccess(c, "Unread count retrieved")
	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
