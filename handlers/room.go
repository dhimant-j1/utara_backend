package handlers

import (
	"context"
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"strings"
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
		Building:        req.Building,
		IsOccupied:      false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		RoomCategoryId:  req.RoomCategoryId,
	}

	result, err := config.DB.Collection("rooms").InsertOne(context.Background(), room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating room"})
		return
	}

	room.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, room)
}

// CreateMultipleRooms creates multiple new rooms
func CreateMultipleRooms(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}

	if !strings.HasSuffix(fileHeader.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only CSV files are allowed"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV headers"})
		return
	}

	var rooms []interface{}
	var skippedRows []string
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		// Map CSV headers to row values
		data := map[string]string{}
		for i, header := range headers {
			data[header] = row[i]
		}

		// Validate RoomType
		roomType := models.RoomType(strings.TrimSpace(data["type"]))
		switch roomType {
		case models.ShreeHariPlus, models.ShreeHari, models.SarjuPlus, models.Sarju, models.NeelkanthPlus, models.Neelkanth:
			// Valid type
		default:
			skippedRows = append(skippedRows, "Invalid room type for room_number: "+data["room_number"])
			continue // Skip this row
		}

		// Parse and validate beds from comma-separated values
		bedsStr := strings.Split(data["beds"], ",")
		var beds []models.Bed
		validBeds := true
		for _, b := range bedsStr {
			trimmedBedType := models.BedType(strings.TrimSpace(b))
			if trimmedBedType == "" {
				continue
			}
			switch trimmedBedType {
			case models.Single, models.Double, models.ExtraBed:
				beds = append(beds, models.Bed{Type: trimmedBedType})
			default:
				validBeds = false
				break // Invalid bed type found
			}
		}

		if !validBeds {
			skippedRows = append(skippedRows, "Invalid bed type for room_number: "+data["room_number"])
			continue // Skip this row
		}

		// Parse primitive values
		floor, _ := strconv.Atoi(data["floor"])
		sofaSetQty, _ := strconv.Atoi(data["sofa_set_quantity"])
		hasAC := strings.ToLower(data["has_ac"]) == "true"
		hasGeyser := strings.ToLower(data["has_geyser"]) == "true"
		hasSofaSet := strings.ToLower(data["has_sofa_set"]) == "true"
		isVisible := strings.ToLower(data["is_visible"]) == "true"

		// Parse images from comma-separated values
		imagesStr := strings.Split(data["images"], ",")
		descStr := strings.Split(data["images_description"], ",")

		var images []models.RoomImage
		for i, imgURL := range imagesStr {
			trimmedURL := strings.TrimSpace(imgURL)
			if trimmedURL == "" {
				continue
			}
			description := ""
			if i < len(descStr) {
				description = strings.TrimSpace(descStr[i])
			}

			images = append(images, models.RoomImage{
				URL:         trimmedURL,
				Description: description,
				UploadedAt:  time.Now(),
			})
		}

		room := models.Room{
			RoomNumber:      data["room_number"],
			Floor:           floor,
			Type:            roomType,
			Beds:            beds,
			HasGeyser:       hasGeyser,
			HasAC:           hasAC,
			HasSofaSet:      hasSofaSet,
			SofaSetQuantity: sofaSetQty,
			ExtraAmenities:  data["extra_amenities"],
			IsVisible:       isVisible,
			IsOccupied:      false,
			Building:        data["building"],
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Check if room already exists
		var existing models.Room
		err = config.DB.Collection("rooms").FindOne(context.Background(), bson.M{
			"room_number": room.RoomNumber,
			"building":    room.Building,
		}).Decode(&existing)

		if errors.Is(err, mongo.ErrNoDocuments) {
			rooms = append(rooms, room)
		} else {
			skippedRows = append(skippedRows, "Room already exists: "+data["room_number"])
		}
	}

	if len(rooms) > 0 {
		_, err = config.DB.Collection("rooms").InsertMany(context.Background(), rooms)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert rooms"})
			return
		}
	}

	response := gin.H{
		"message":         "Room import process finished",
		"imported_rooms":  len(rooms),
		"skipped_rows":    len(skippedRows),
		"skipped_details": skippedRows,
	}

	if len(rooms) == 0 && len(skippedRows) > 0 {
		c.JSON(http.StatusConflict, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRooms returns all rooms with optional filters
func GetRooms(c *gin.Context) {
	filter := bson.M{}

	// Apply filters if provided
	if floor := c.Query("floor"); floor != "" {
		// floor will be int, so convert string to int
		floorInt, err := strconv.Atoi(floor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid floor value"})
			return
		}
		filter["floor"] = floorInt
	}
	if roomType := c.Query("type"); roomType != "" {
		filter["type"] = roomType
	}
	if building := c.Query("building"); building != "" {
		filter["building"] = building
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

// DeleteRoom deletes a room by ID
func DeleteRoom(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	result, err := config.DB.Collection("rooms").DeleteOne(
		context.Background(),
		bson.M{"_id": id},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting room"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}

func CreateRoomCategory(c *gin.Context) {
	var req models.RoomCategory

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = primitive.NewObjectID()
	req.CreatedAt = time.Now()

	_, err := config.DB.Collection("room_category").InsertOne(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Room category created successfully",
		"category": req,
	})
}

func GetRoomCategories(c *gin.Context) {
	cursor, err := config.DB.Collection("room_category").Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching categories"})
		return
	}
	defer cursor.Close(context.Background())

	var categories []models.RoomCategory
	if err := cursor.All(context.Background(), &categories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func UpdateRoomCategory(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req["updated_at"] = time.Now()

	update := bson.M{"$set": req}

	res, err := config.DB.Collection("room_category").UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating category"})
		return
	}

	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room category updated successfully"})
}

func DeleteRoomCategory(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room category ID"})
		return
	}

	result, err := config.DB.Collection("room_categories").DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting room category"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room category not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room category deleted successfully"})
}

// GetBuildings returns a list of unique buildings
func GetBuildings(c *gin.Context) {
	// Use MongoDB aggregation to get distinct buildings
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"building":   bson.M{"$ne": ""},
				"is_visible": true,
			},
		},
		{
			"$group": bson.M{
				"_id": "$building",
			},
		},
		{
			"$sort": bson.M{
				"_id": 1,
			},
		},
	}

	cursor, err := config.DB.Collection("rooms").Aggregate(context.Background(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching buildings"})
		return
	}
	defer cursor.Close(context.Background())

	var buildings []string
	for cursor.Next(context.Background()) {
		var result struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		if result.ID != "" {
			buildings = append(buildings, result.ID)
		}
	}

	if buildings == nil {
		buildings = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"buildings": buildings,
		"count":     len(buildings),
	})
}

// GetFloors returns a list of floors for a specific building
func GetFloors(c *gin.Context) {
	building := c.Query("building")
	if building == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Building parameter is required"})
		return
	}

	// Use MongoDB aggregation to get distinct floors for the specified building
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"building":   building,
				"is_visible": true,
			},
		},
		{
			"$group": bson.M{
				"_id": "$floor",
			},
		},
		{
			"$sort": bson.M{
				"_id": 1,
			},
		},
	}

	cursor, err := config.DB.Collection("rooms").Aggregate(context.Background(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching floors"})
		return
	}
	defer cursor.Close(context.Background())

	var floors []int
	for cursor.Next(context.Background()) {
		var result struct {
			ID int `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		floors = append(floors, result.ID)
	}

	if floors == nil {
		floors = []int{}
	}

	c.JSON(http.StatusOK, gin.H{
		"building": building,
		"floors":   floors,
		"count":    len(floors),
	})
}
