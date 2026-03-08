package controllers

import (
	"net/http"
	"rest-api/models"
	"rest-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(db *gorm.DB) {
	DB = db
	DB.AutoMigrate(&models.User{})
}

func GetUsers(c *gin.Context) {
	var users []models.User
	DB.Find(&users)
	utils.LogSuccess(c, "Users retrieved successfully")
	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := DB.First(&user, id).Error; err != nil {
		utils.LogError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	utils.LogSuccess(c, "User retrieved successfully")
	c.JSON(http.StatusOK, user)
}

func CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	if err := utils.ValidateStruct(&user); err != nil {
		utils.LogError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	var existingUser models.User
	if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		utils.LogError(c, http.StatusConflict, "EMAIL_EXISTS", "Email already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	DB.Create(&user)
	utils.LogSuccess(c, "User created successfully")
	c.JSON(http.StatusCreated, user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := DB.First(&user, id).Error; err != nil {
		utils.LogError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	if err := utils.ValidateStruct(&input); err != nil {
		utils.LogError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed")
		c.JSON(http.StatusBadRequest, utils.ValidationError(err))
		return
	}

	DB.Model(&user).Updates(input)
	utils.LogSuccess(c, "User updated successfully")
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := DB.First(&user, id).Error; err != nil {
		utils.LogError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	DB.Delete(&user)
	utils.LogSuccess(c, "User deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
