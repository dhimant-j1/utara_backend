package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"utara_backend/config"
	"utara_backend/models"
)

// CreateRoomRequest creates a new room request
func CreateRoomRequest(c *gin.Context) {
	var req models.CreateRoomRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}

	roomRequest := models.RoomRequest{
		UserID:       userObjID,
		Name:         user.Name,
		Place:        req.Place,
		Purpose:      req.Purpose,
		CheckInDate:  req.CheckInDate,
		CheckOutDate: req.CheckOutDate,
		NumberOfPeople: models.PeopleCount{
			Male:     req.NumberOfPeople.Male,
			Female:   req.NumberOfPeople.Female,
			Children: req.NumberOfPeople.Children,
			Total:    req.NumberOfPeople.Male + req.NumberOfPeople.Female + req.NumberOfPeople.Children,
		},
		//PreferredType:   req.PreferredType,
		SpecialRequests: req.SpecialRequests,
		Status:          models.StatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result, err := config.DB.Collection("room_requests").InsertOne(context.Background(), roomRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating room request"})
		return
	}

	roomRequest.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, roomRequest)
}

// GetRoomRequests returns all room requests with optional filters and room details
func GetRoomRequests(c *gin.Context) {
	filter := bson.M{}

	// Get user role from context
	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}

	// Regular users can only see their own requests
	if user.Role == models.RoleUser {
		filter["user_id"] = userObjID
	} else {
		// Apply filters if provided (for SUPER_ADMIN and STAFF)
		if status := c.Query("status"); status != "" {
			filter["status"] = status
		}
		if filterUserID := c.Query("user_id"); filterUserID != "" {
			userObjID, err := primitive.ObjectIDFromHex(filterUserID)
			if err == nil {
				filter["user_id"] = userObjID
			}
		}
	}

	cursor, err := config.DB.Collection("room_requests").Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching room requests"})
		return
	}
	defer cursor.Close(context.Background())

	var requests []models.RoomRequest
	if err := cursor.All(context.Background(), &requests); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding room requests"})
		return
	}

	// Create a response struct to include room details
	type RoomRequestWithDetails struct {
		models.RoomRequest
		User       *models.User           `json:"user,omitempty"`
		Room       *models.Room           `json:"room,omitempty"`
		Assignment *models.RoomAssignment `json:"assignment,omitempty"`
	}

	var requestsWithDetails []RoomRequestWithDetails

	// For each request, find its assignment and room (if any)
	for _, request := range requests {
		// get userDetails
		var userDetails models.User
		err := config.DB.Collection("users").FindOne(
			context.Background(),
			bson.M{"_id": request.UserID},
		).Decode(&userDetails)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user details"})
			return
		}

		requestWithDetails := RoomRequestWithDetails{
			RoomRequest: request,
			User:        &userDetails,
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

		requestsWithDetails = append(requestsWithDetails, requestWithDetails)
	}

	c.JSON(http.StatusOK, requestsWithDetails)
}

// ProcessRoomRequest processes (approve/reject) a room request
func ProcessRoomRequest(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var req models.ProcessRoomRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffID, _ := c.Get("user_id")
	staffObjID, _ := primitive.ObjectIDFromHex(staffID.(string))

	update := bson.M{
		"$set": bson.M{
			"status":       req.Status,
			"processed_by": staffObjID,
			"processed_at": time.Now(),
			"updated_at":   time.Now(),
		},
	}

	var roomRequest models.RoomRequest
	err = config.DB.Collection("room_requests").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&roomRequest)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing room request"})
		return
	}

	// If request is approved and room ID is provided, create room assignment
	if req.Status == models.StatusApproved && req.RoomID != nil {
		assignment := models.RoomAssignment{
			RoomID:       *req.RoomID,
			UserID:       roomRequest.UserID,
			RequestID:    roomRequest.ID,
			CheckInDate:  roomRequest.CheckInDate,
			CheckOutDate: roomRequest.CheckOutDate,
			AssignedBy:   staffObjID,
			AssignedAt:   time.Now(),
			CheckedIn:    false,
			CheckedOut:   false,
		}

		_, err := config.DB.Collection("room_assignments").InsertOne(context.Background(), assignment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating room assignment"})
			return
		}

		// Update room status to occupied
		_, err = config.DB.Collection("rooms").UpdateOne(
			context.Background(),
			bson.M{"_id": req.RoomID},
			bson.M{"$set": bson.M{"is_occupied": true}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating room status"})
			return
		}
	}

	c.JSON(http.StatusOK, roomRequest)
}

// AssignRoom assigns a room to a user
func AssignRoom(c *gin.Context) {
	var req models.RoomAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffID, _ := c.Get("user_id")
	staffObjID, _ := primitive.ObjectIDFromHex(staffID.(string))

	// Check if room is available
	var room models.Room
	err := config.DB.Collection("rooms").FindOne(context.Background(), bson.M{
		"_id":         req.RoomID,
		"is_occupied": false,
	}).Decode(&room)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Room is not available"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking room availability"})
		return
	}

	assignment := models.RoomAssignment{
		RoomID:       req.RoomID,
		UserID:       req.UserID,
		RequestID:    req.RequestID,
		CheckInDate:  req.CheckInDate,
		CheckOutDate: req.CheckOutDate,
		AssignedBy:   staffObjID,
		AssignedAt:   time.Now(),
		CheckedIn:    false,
		CheckedOut:   false,
	}

	result, err := config.DB.Collection("room_assignments").InsertOne(context.Background(), assignment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating room assignment"})
		return
	}

	// Update room status to occupied
	_, err = config.DB.Collection("rooms").UpdateOne(
		context.Background(),
		bson.M{"_id": req.RoomID},
		bson.M{"$set": bson.M{"is_occupied": true}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating room status"})
		return
	}

	assignment.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, assignment)
}

// CheckInRoom marks a room assignment as checked in
func CheckInRoom(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"checked_in":    true,
			"checked_in_at": now,
		},
	}

	var assignment models.RoomAssignment
	err = config.DB.Collection("room_assignments").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id, "checked_in": false},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&assignment)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room assignment not found or already checked in"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking in"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}

// CheckOutRoom marks a room assignment as checked out
func CheckOutRoom(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"checked_out":    true,
			"checked_out_at": now,
		},
	}

	var assignment models.RoomAssignment
	err = config.DB.Collection("room_assignments").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id, "checked_in": true, "checked_out": false},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&assignment)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room assignment not found, not checked in, or already checked out"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking out"})
		return
	}

	// Update room status to available
	_, err = config.DB.Collection("rooms").UpdateOne(
		context.Background(),
		bson.M{"_id": assignment.RoomID},
		bson.M{"$set": bson.M{"is_occupied": false}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating room status"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}
