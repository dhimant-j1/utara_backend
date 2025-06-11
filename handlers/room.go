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

// CreateRoom creates a new room
func CreateRoom(c *gin.Context) {
	var req models.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if room number already exists
	var existingRoom models.Room
	err := config.DB.Collection("rooms").FindOne(context.Background(), bson.M{"room_number": req.RoomNumber}).Decode(&existingRoom)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room number already exists"})
		return
	}

	room := models.Room{
		RoomNumber:      req.RoomNumber,
		Floor:           req.Floor,
		Type:            req.Type,
		Beds:            req.Beds,
		HasGeyser:       req.HasGeyser,
		HasAC:           req.HasAC,
		HasSofaSet:      req.HasSofaSet,
		SofaSetQuantity: req.SofaSetQuantity,
		ExtraAmenities:  req.ExtraAmenities,
		IsVisible:       req.IsVisible,
		Images:          req.Images,
		IsOccupied:      false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result, err := config.DB.Collection("rooms").InsertOne(context.Background(), room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating room"})
		return
	}

	room.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, room)
}

// GetRooms returns all rooms with optional filters
func GetRooms(c *gin.Context) {
	filter := bson.M{}

	// Apply filters if provided
	if floor := c.Query("floor"); floor != "" {
		filter["floor"] = floor
	}
	if roomType := c.Query("type"); roomType != "" {
		filter["type"] = roomType
	}
	if isVisible := c.Query("is_visible"); isVisible != "" {
		filter["is_visible"] = isVisible == "true"
	}
	if isOccupied := c.Query("is_occupied"); isOccupied != "" {
		filter["is_occupied"] = isOccupied == "true"
	}

	// Get user role from context
	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}

	// Regular users can only see visible rooms
	if user.Role == models.RoleUser {
		filter["is_visible"] = true
	}

	cursor, err := config.DB.Collection("rooms").Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching rooms"})
		return
	}
	defer cursor.Close(context.Background())

	var rooms = []models.Room{}
	if err := cursor.All(context.Background(), &rooms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding rooms"})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

// GetRoom returns a specific room by ID
func GetRoom(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	// Get user role from context
	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
	var user models.User
	err = config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}

	// Build filter based on role
	filter := bson.M{"_id": id}
	if user.Role == models.RoleUser {
		filter["is_visible"] = true
	}

	var room models.Room
	err = config.DB.Collection("rooms").FindOne(context.Background(), filter).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found or not accessible"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching room"})
		return
	}

	c.JSON(http.StatusOK, room)
}

// UpdateRoom updates a room by ID
func UpdateRoom(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var req models.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.RoomNumber != nil {
		// Check if room number already exists
		var existingRoom models.Room
		err := config.DB.Collection("rooms").FindOne(context.Background(), bson.M{
			"room_number": req.RoomNumber,
			"_id":         bson.M{"$ne": id},
		}).Decode(&existingRoom)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Room number already exists"})
			return
		}
		update["$set"].(bson.M)["room_number"] = *req.RoomNumber
	}

	if req.Floor != nil {
		update["$set"].(bson.M)["floor"] = *req.Floor
	}
	if req.Type != nil {
		update["$set"].(bson.M)["type"] = *req.Type
	}
	if req.Beds != nil {
		update["$set"].(bson.M)["beds"] = req.Beds
	}
	if req.HasGeyser != nil {
		update["$set"].(bson.M)["has_geyser"] = *req.HasGeyser
	}
	if req.HasAC != nil {
		update["$set"].(bson.M)["has_ac"] = *req.HasAC
	}
	if req.HasSofaSet != nil {
		update["$set"].(bson.M)["has_sofa_set"] = *req.HasSofaSet
	}
	if req.SofaSetQuantity != nil {
		update["$set"].(bson.M)["sofa_set_quantity"] = *req.SofaSetQuantity
	}
	if req.ExtraAmenities != nil {
		update["$set"].(bson.M)["extra_amenities"] = *req.ExtraAmenities
	}
	if req.IsVisible != nil {
		update["$set"].(bson.M)["is_visible"] = *req.IsVisible
	}
	if req.Images != nil {
		update["$set"].(bson.M)["images"] = req.Images
	}

	result := config.DB.Collection("rooms").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var updatedRoom models.Room
	if err := result.Decode(&updatedRoom); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating room"})
		return
	}

	c.JSON(http.StatusOK, updatedRoom)
}

// GetRoomStats returns statistics about rooms
func GetRoomStats(c *gin.Context) {
	totalRooms, err := config.DB.Collection("rooms").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching room stats"})
		return
	}

	occupiedRooms, err := config.DB.Collection("rooms").CountDocuments(context.Background(), bson.M{"is_occupied": true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching room stats"})
		return
	}

	stats := gin.H{
		"total_rooms":     totalRooms,
		"occupied_rooms":  occupiedRooms,
		"available_rooms": totalRooms - occupiedRooms,
	}

	c.JSON(http.StatusOK, stats)
}
