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

				if req.DiningHall != "" {
					var category models.FoodPassCategory
					err := config.DB.Collection("food_pass_categories").FindOne(context.Background(), bson.M{
						"building_name": req.DiningHall,
					}).Decode(&category)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching food pass category"})
						return
					}
					req.ColorCode = category.ColorCode
				}

				pass := models.FoodPass{
					ID:         id,
					UserID:     req.UserID,
					MemberName: memberName,
					MealType:   mealType,
					Date:       currentDate,
					QRCode:     qrCode,
					IsUsed:     false,
					DiningHall: req.DiningHall,
					ColorCode:  req.ColorCode,
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

func UpdateFoodPass(c *gin.Context) {

	passID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food pass ID"})
		return
	}

	var req struct {
		MemberName *string          `json:"member_name,omitempty"`
		MealType   *models.MealType `json:"meal_type,omitempty"`
		Date       *time.Time       `json:"date,omitempty"`
		IsUsed     *bool            `json:"is_used,omitempty"`
		DiningHall *string          `json:"dining_hall,omitempty"`
		ColorCode  *string          `json:"color_code,omitempty"`
		UsedAt     *time.Time       `json:"used_at,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{"updated_at": time.Now()}
	set := bson.M{}
	if req.MemberName != nil {
		set["member_name"] = *req.MemberName
	}
	if req.MealType != nil {
		set["meal_type"] = *req.MealType
	}
	if req.Date != nil {
		set["date"] = *req.Date
	}
	if req.IsUsed != nil {
		set["is_used"] = *req.IsUsed
	}
	if req.DiningHall != nil {
		set["dining_hall"] = *req.DiningHall
	}
	if req.ColorCode != nil {
		set["color_code"] = *req.ColorCode
	}
	if req.UsedAt != nil {
		set["used_at"] = req.UsedAt
	}

	if len(set) > 0 {
		update["$set"] = set
	}

	result := config.DB.Collection("food_passes").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": passID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var updatedPass models.FoodPass
	if err := result.Decode(&updatedPass); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Food pass not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating food pass"})
		return
	}

	c.JSON(http.StatusOK, updatedPass)
}

// CreateFoodPassCategory creates a new food pass category
func CreateFoodPassCategory(c *gin.Context) {
	var req models.FoodPassCategory

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate color code (simple check: must start with # and length 7)
	if len(req.ColorCode) != 7 || req.ColorCode[0] != '#' {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid color code. Use hex format like #FF5733"})
		return
	}

	req.ID = primitive.NewObjectID()
	req.CreatedAt = time.Now()

	_, err := config.DB.Collection("food_pass_categories").InsertOne(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Food pass category created successfully",
		"category": req,
	})
}

// GetFoodPassCategories lists all categories
func GetFoodPassCategories(c *gin.Context) {
	cursor, err := config.DB.Collection("food_pass_categories").Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching categories"})
		return
	}
	defer cursor.Close(context.Background())

	var categories []models.FoodPassCategory = []models.FoodPassCategory{}
	if err := cursor.All(context.Background(), &categories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// UpdateFoodPassCategory updates building name or color code
func UpdateFoodPassCategory(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var req struct {
		BuildingName string `json:"building_name"`
		ColorCode    string `json:"color_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{}
	if req.BuildingName != "" {
		update["building_name"] = req.BuildingName
	}
	if req.ColorCode != "" {
		if len(req.ColorCode) != 7 || req.ColorCode[0] != '#' {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid color code. Use hex format like #FF5733"})
			return
		}
		update["color_code"] = req.ColorCode
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	_, err = config.DB.Collection("food_pass_categories").UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully"})
}

// DeleteFoodPassCategory deletes a category by ID
func DeleteFoodPassCategory(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	_, err = config.DB.Collection("food_pass_categories").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
