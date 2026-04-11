package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"utara_backend/config"
	"utara_backend/models"
)

// GetUserTypeConfigs returns all user type configurations
func GetUserTypeConfigs(c *gin.Context) {
	cursor, err := config.DB.Collection("user_type_configs").Find(
		context.Background(),
		bson.M{},
		options.Find().SetSort(bson.D{{Key: "user_type", Value: 1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user type configs"})
		return
	}
	defer cursor.Close(context.Background())

	var configs []models.UserTypeConfig
	if err := cursor.All(context.Background(), &configs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user type configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"count":   len(configs),
	})
}

// GetUserTypeConfig returns config for a specific user type
func GetUserTypeConfig(c *gin.Context) {
	userType := c.Param("type")

	var userTypeConfig models.UserTypeConfig
	err := config.DB.Collection("user_type_configs").FindOne(
		context.Background(),
		bson.M{"user_type": userType},
	).Decode(&userTypeConfig)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User type config not found"})
		return
	}

	c.JSON(http.StatusOK, userTypeConfig)
}

// CreateUserTypeConfig creates a new user type configuration
func CreateUserTypeConfig(c *gin.Context) {
	var req struct {
		UserType      string `json:"user_type" binding:"required"`
		DepositAmount int    `json:"deposit_amount"`
		IsFOC         bool   `json:"is_foc"`
		Description   string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user type already exists
	var existing models.UserTypeConfig
	err := config.DB.Collection("user_type_configs").FindOne(
		context.Background(),
		bson.M{"user_type": req.UserType},
	).Decode(&existing)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User type config already exists"})
		return
	}

	newConfig := models.UserTypeConfig{
		UserType:      req.UserType,
		DepositAmount: req.DepositAmount,
		IsFOC:         req.IsFOC,
		Description:   req.Description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	result, err := config.DB.Collection("user_type_configs").InsertOne(context.Background(), newConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user type config"})
		return
	}

	newConfig.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, newConfig)
}

// UpdateUserTypeConfig updates a user type configuration
func UpdateUserTypeConfig(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		DepositAmount *int    `json:"deposit_amount,omitempty"`
		IsFOC         *bool   `json:"is_foc,omitempty"`
		Description   *string `json:"description,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.DepositAmount != nil {
		update["$set"].(bson.M)["deposit_amount"] = *req.DepositAmount
	}
	if req.IsFOC != nil {
		update["$set"].(bson.M)["is_foc"] = *req.IsFOC
	}
	if req.Description != nil {
		update["$set"].(bson.M)["description"] = *req.Description
	}

	var updatedConfig models.UserTypeConfig
	err = config.DB.Collection("user_type_configs").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedConfig)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user type config"})
		return
	}

	c.JSON(http.StatusOK, updatedConfig)
}

// DeleteUserTypeConfig deletes a user type configuration
func DeleteUserTypeConfig(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := config.DB.Collection("user_type_configs").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user type config"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User type config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User type config deleted successfully"})
}

// GetDepositForUser returns the deposit amount required for a user
func GetDepositForUser(userType string) (amount int, isFOC bool) {
	var userTypeConfig models.UserTypeConfig
	err := config.DB.Collection("user_type_configs").FindOne(
		context.Background(),
		bson.M{"user_type": userType},
	).Decode(&userTypeConfig)

	if err == nil {
		if userTypeConfig.IsFOC {
			return 0, true
		}
		return userTypeConfig.DepositAmount, false
	}

	// Default fallback: Plus type customers get free of cost
	if strings.Contains(strings.ToLower(userType), "plus") {
		return 0, true
	}

	// Otherwise, fallback to the deposit amount defined for the equivalent Room Type
	amount = GetDepositAmountForRoomType(models.RoomType(userType))
	return amount, false
}

// CheckUserRequiresDeposit checks if a user needs to pay deposit based on their user type
func CheckUserRequiresDeposit(c *gin.Context) {
	userType := c.Query("user_type")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_type is required"})
		return
	}

	amount, isFOC := GetDepositForUser(userType)

	c.JSON(http.StatusOK, gin.H{
		"user_type":        userType,
		"requires_deposit": !isFOC && amount > 0,
		"deposit_amount":   amount,
		"is_foc":           isFOC,
		"currency":         "INR",
	})
}
