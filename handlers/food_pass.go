package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"utara_backend/config"
	"utara_backend/models"
	"utara_backend/utils"
)

// GenerateFoodPasses generates food passes for a user and their family members
func GenerateFoodPasses(c *gin.Context) {
	var req models.GenerateFoodPassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffID, _ := c.Get("user_id")
	staffObjID, _ := primitive.ObjectIDFromHex(staffID.(string))

	// Check if user exists and is active
	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{
		"_id": req.UserID,
	}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}

	var passes []models.FoodPass
	currentDate := req.StartDate

	// Generate passes for each day between start and end date
	for !currentDate.After(req.EndDate) {
		// Generate passes for each meal type
		mealTypes := []models.MealType{models.Breakfast, models.Lunch, models.Dinner}

		for _, memberName := range req.MemberNames {
			for _, mealType := range mealTypes {
				id := primitive.NewObjectID()
				// Generate QR code data

				qrCode, err := utils.GenerateQRCode(id.Hex())
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating QR code"})
					return
				}

				pass := models.FoodPass{
					ID:         id,
					UserID:     req.UserID,
					MemberName: memberName,
					MealType:   mealType,
					Date:       currentDate,
					QRCode:     qrCode,
					IsUsed:     false,
					CreatedBy:  staffObjID,
					CreatedAt:  time.Now(),
				}
				passes = append(passes, pass)
			}
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Insert all passes
	var passInterfaces []interface{}
	for _, pass := range passes {
		passInterfaces = append(passInterfaces, pass)
	}

	_, err = config.DB.Collection("food_passes").InsertMany(context.Background(), passInterfaces)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating food passes"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Food passes generated successfully",
		"total_passes": len(passes),
	})
}

// GetUserFoodPasses returns food passes for a specific user
func GetUserFoodPasses(c *gin.Context) {
	targetUserID, err := primitive.ObjectIDFromHex(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current user's role from context
	currentUserID, _ := c.Get("user_id")
	currentUserObjID, _ := primitive.ObjectIDFromHex(currentUserID.(string))
	var currentUser models.User
	err = config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": currentUserObjID}).Decode(&currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
		return
	}

	// Regular users can only see their own passes
	if currentUser.Role == models.RoleUser && currentUserObjID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own food passes"})
		return
	}

	filter := bson.M{"user_id": targetUserID}

	// Apply additional filters if provided
	if date := c.Query("date"); date != "" {
		parsedDate, err := time.Parse("2006-01-02", date)
		if err == nil {
			filter["date"] = bson.M{
				"$gte": parsedDate,
				"$lt":  parsedDate.Add(24 * time.Hour),
			}
		}
	}
	if isUsed := c.Query("is_used"); isUsed != "" {
		filter["is_used"] = isUsed == "true"
	}

	cursor, err := config.DB.Collection("food_passes").Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching food passes"})
		return
	}
	defer cursor.Close(context.Background())

	var passes []models.FoodPass
	if err := cursor.All(context.Background(), &passes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding food passes"})
		return
	}

	c.JSON(http.StatusOK, passes)
}

// ScanFoodPass marks a food pass as used
func ScanFoodPass(c *gin.Context) {
	var req models.ScanFoodPassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_used": true,
			"used_at": now,
		},
	}

	var pass models.FoodPass
	err := config.DB.Collection("food_passes").FindOneAndUpdate(
		context.Background(),
		bson.M{
			"_id":     req.PassID,
			"is_used": false,
			"date":    bson.M{"$gte": now.Truncate(24 * time.Hour)}, // Only allow scanning passes for today or future
		},
		update,
	).Decode(&pass)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Food pass not found, already used, or expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning food pass"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Food pass scanned successfully"})
}
