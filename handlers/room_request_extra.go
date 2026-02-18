package handlers

import (
	"context"
	"net/http"
	"time"

	"utara_backend/config"
	"utara_backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateRoomRequest updates a room request (only number of people) for pending requests
func UpdateRoomRequest(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var req models.PeopleCount
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))

	// Calculate total
	req.Total = req.Male + req.Female + req.Children

	filter := bson.M{
		"_id":     id,
		"user_id": userObjID,
		"status":  models.StatusPending,
	}

	update := bson.M{
		"$set": bson.M{
			"number_of_people": req,
			"updated_at":       time.Now(),
		},
	}

	result, err := config.DB.Collection("room_requests").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating room request"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room request not found or cannot be updated (must be pending)"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room request updated successfully"})
}

// DeleteRoomRequest deletes a pending room request
func DeleteRoomRequest(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))

	filter := bson.M{
		"_id":     id,
		"user_id": userObjID,
		"status":  models.StatusPending,
	}

	result, err := config.DB.Collection("room_requests").DeleteOne(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting room request"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room request not found or cannot be deleted (must be pending)"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room request deleted successfully"})
}
