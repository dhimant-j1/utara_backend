package handlers

import (
	"context"
	"net/http"

	"utara_backend/config"
	"utara_backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetRoomRequestByID returns a single room request by ID
func GetRoomRequestByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))

	// Get user role to determine access
	var user models.User
	err = config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}

	filter := bson.M{"_id": id}
	// If regular user, ensure they own the request
	if user.Role == models.RoleUser {
		filter["user_id"] = userObjID
	}

	var request models.RoomRequest
	err = config.DB.Collection("room_requests").FindOne(context.Background(), filter).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching room request"})
		return
	}

	type RoomRequestWithDetails struct {
		models.RoomRequest
		User       *models.User           `json:"user,omitempty"`
		Room       *models.Room           `json:"room,omitempty"`
		Assignment *models.RoomAssignment `json:"assignment,omitempty"`
	}

	// Get user details
	var userDetails models.User
	err = config.DB.Collection("users").FindOne(
		context.Background(),
		bson.M{"_id": request.UserID},
	).Decode(&userDetails)

	requestWithDetails := RoomRequestWithDetails{
		RoomRequest: request,
	}

	if err == nil {
		requestWithDetails.User = &userDetails
	}

	// Find assignment for this request
	var assignment models.RoomAssignment
	err = config.DB.Collection("room_assignments").FindOne(
		context.Background(),
		bson.M{"request_id": request.ID},
	).Decode(&assignment)

	if err == nil {
		requestWithDetails.Assignment = &assignment

		// If assignment exists, get room details
		var room models.Room
		err = config.DB.Collection("rooms").FindOne(
			context.Background(),
			bson.M{"_id": assignment.RoomID},
		).Decode(&room)

		if err == nil {
			requestWithDetails.Room = &room
		}
	}

	c.JSON(http.StatusOK, requestWithDetails)
}
