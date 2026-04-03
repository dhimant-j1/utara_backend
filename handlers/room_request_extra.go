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

// AdminUpdateRoomRequest allows admin to update any room request (including approved ones)
func AdminUpdateRoomRequest(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var req struct {
		CheckInDate     *time.Time          `json:"check_in_date"`
		CheckOutDate    *time.Time          `json:"check_out_date"`
		NumberOfPeople  *models.PeopleCount `json:"number_of_people"`
		Place           *string             `json:"place"`
		Purpose         *string             `json:"purpose"`
		Reference       *string             `json:"reference"`
		SpecialRequests *string             `json:"special_requests"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update document with only provided fields
	updateFields := bson.M{
		"updated_at": time.Now(),
	}

	if req.CheckInDate != nil {
		updateFields["check_in_date"] = *req.CheckInDate
	}
	if req.CheckOutDate != nil {
		updateFields["check_out_date"] = *req.CheckOutDate
	}
	if req.NumberOfPeople != nil {
		req.NumberOfPeople.Total = req.NumberOfPeople.Male + req.NumberOfPeople.Female + req.NumberOfPeople.Children
		updateFields["number_of_people"] = *req.NumberOfPeople
	}
	if req.Place != nil {
		updateFields["place"] = *req.Place
	}
	if req.Purpose != nil {
		updateFields["purpose"] = *req.Purpose
	}
	if req.Reference != nil {
		updateFields["reference"] = *req.Reference
	}
	if req.SpecialRequests != nil {
		updateFields["special_requests"] = *req.SpecialRequests
	}

	filter := bson.M{"_id": id}
	update := bson.M{"$set": updateFields}

	result, err := config.DB.Collection("room_requests").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating room request"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room request not found"})
		return
	}

	// Fetch and return the updated document
	var updatedRequest models.RoomRequest
	err = config.DB.Collection("room_requests").FindOne(context.Background(), filter).Decode(&updatedRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Room request updated successfully"})
		return
	}

	c.JSON(http.StatusOK, updatedRequest)
}
