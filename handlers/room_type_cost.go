package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"utara_backend/config"
	"utara_backend/models"
)

// GetRoomTypeCosts returns all room type deposit costs
func GetRoomTypeCosts(c *gin.Context) {
	cursor, err := config.DB.Collection("room_type_costs").Find(
		context.Background(),
		bson.M{},
		options.Find().SetSort(bson.D{{Key: "room_type", Value: 1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room type costs"})
		return
	}
	defer cursor.Close(context.Background())

	var costs []models.RoomTypeCost
	if err := cursor.All(context.Background(), &costs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode room type costs"})
		return
	}

	// If no costs are set, return defaults
	if len(costs) == 0 {
		costs = models.GetDefaultRoomTypeCosts()
	}

	c.JSON(http.StatusOK, gin.H{
		"costs": costs,
		"count": len(costs),
	})
}

// GetRoomTypeCost returns cost for a specific room type
func GetRoomTypeCost(c *gin.Context) {
	roomType := models.RoomType(c.Param("type"))

	var cost models.RoomTypeCost
	err := config.DB.Collection("room_type_costs").FindOne(
		context.Background(),
		bson.M{"room_type": roomType},
	).Decode(&cost)

	if err != nil {
		// Return default if not found
		defaults := models.GetDefaultRoomTypeCosts()
		for _, d := range defaults {
			if d.RoomType == roomType {
				c.JSON(http.StatusOK, d)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Room type not found"})
		return
	}

	c.JSON(http.StatusOK, cost)
}

// UpdateRoomTypeCost updates the deposit amount for a room type
func UpdateRoomTypeCost(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		DepositAmount int    `json:"deposit_amount"`
		Currency      string `json:"currency"`
		Description   string `json:"description"`
		IsActive      *bool  `json:"is_active,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"deposit_amount": req.DepositAmount,
			"updated_at":     time.Now(),
		},
	}

	if req.Currency != "" {
		update["$set"].(bson.M)["currency"] = req.Currency
	}
	if req.Description != "" {
		update["$set"].(bson.M)["description"] = req.Description
	}
	if req.IsActive != nil {
		update["$set"].(bson.M)["is_active"] = *req.IsActive
	}

	var cost models.RoomTypeCost
	err = config.DB.Collection("room_type_costs").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&cost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room type cost"})
		return
	}

	c.JSON(http.StatusOK, cost)
}

// InitializeRoomTypeCosts creates default room type costs if none exist
func InitializeRoomTypeCosts(c *gin.Context) {
	// Check if any costs exist
	count, err := config.DB.Collection("room_type_costs").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing costs"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Room type costs already initialized"})
		return
	}

	// Insert defaults
	defaults := models.GetDefaultRoomTypeCosts()
	docs := make([]interface{}, len(defaults))
	for i, d := range defaults {
		docs[i] = d
	}

	_, err = config.DB.Collection("room_type_costs").InsertMany(context.Background(), docs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize room type costs"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Room type costs initialized successfully"})
}

// GetDepositAmountForRoomType returns the deposit amount for a room type
func GetDepositAmountForRoomType(roomType models.RoomType) int {
	var cost models.RoomTypeCost
	err := config.DB.Collection("room_type_costs").FindOne(
		context.Background(),
		bson.M{"room_type": roomType, "is_active": true},
	).Decode(&cost)

	if err != nil {
		// Return default amount
		defaults := models.GetDefaultRoomTypeCosts()
		for _, d := range defaults {
			if d.RoomType == roomType {
				return d.DepositAmount
			}
		}
		return 30000 // Default fallback: ₹300
	}

	return cost.DepositAmount
}
