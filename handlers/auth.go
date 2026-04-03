package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"utara_backend/config"
	"utara_backend/models"
)

func GenerateOTP() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "123456" // fallback
	}
	otp := ""
	for i := 0; i < 6; i++ {
		otp += fmt.Sprintf("%d", int(b[i])%10)
	}
	return "123456"
}

// func Signup(c *gin.Context) {
// 	var req models.SignupRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Check if user already exists
// 	var existingUser models.User
// 	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
// 	if err == nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
// 		return
// 	}

// 	// Validate role
// 	if req.Role != models.RoleSuperAdmin && req.Role != models.RoleStaff && req.Role != models.RoleUser {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
// 		return
// 	}

// 	// Check if this is the first user being created
// 	count, err := config.DB.Collection("users").CountDocuments(context.Background(), bson.M{})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user count"})
// 		return
// 	}

// 	// If this is not the first user, apply the regular role checks
// 	if count > 0 {
// 		if req.Role == models.RoleSuperAdmin || req.Role == models.RoleStaff {
// 			userID, exists := c.Get("user_id")
// 			if !exists {
// 				c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin can create admin or staff users"})
// 				return
// 			}

// 			var currentUser models.User
// 			userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
// 			err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&currentUser)
// 			if err != nil || currentUser.Role != models.RoleSuperAdmin {
// 				c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin can create admin or staff users"})
// 				return
// 			}
// 		}
// 	}

// 	// Hash password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
// 		return
// 	}

// 	parts := strings.Split(req.Email, "@")
// 	username := ""
// 	if len(parts) > 0 {
// 		username = parts[0]
// 		fmt.Println("Username:", username)
// 	} else {
// 		fmt.Println("Invalid email format")
// 	}

// 	// Create user
// 	user := models.User{
// 		Email:       req.Email,
// 		UserName:    username,
// 		Password:    string(hashedPassword),
// 		Name:        req.Name,
// 		Role:        req.Role,
// 		PhoneNumber: req.PhoneNumber,
// 		UserType:    "Neelkanth",
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 	}

// 	result, err := config.DB.Collection("users").InsertOne(context.Background(), user)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
// 		return
// 	}

// 	user.ID = result.InsertedID.(primitive.ObjectID)
// 	user.Password = "" // Remove password from response

// 	// Generate JWT token
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"user_id": user.ID.Hex(),
// 		"role":    user.Role, // Add role to JWT claims
// 		"exp":     time.Now().Add(time.Hour * 24).Unix(),
// 	})

// 	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, models.AuthResponse{
// 		Token: tokenString,
// 		User:  user,
// 	})
// }

func Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	// Delete all existing signup OTP entries for this phone number
	_, _ = config.DB.Collection("signup_otps").DeleteMany(context.Background(),
		bson.M{"phone_number": req.PhoneNumber})

	otp := GenerateOTP()
	otpExpiry := time.Now().Add(5 * time.Minute)

	signupEntry := models.SignupOtpEntry{
		PhoneNumber: req.PhoneNumber,
		Request:     req,
		Otp:         otp,
		OtpExpiry:   otpExpiry,
		CreatedAt:   time.Now(),
	}

	_, err = config.DB.Collection("signup_otps").InsertOne(context.Background(), signupEntry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving OTP"})
		return
	}

	/*err = SendWhatsAppMessage(req.PhoneNumber, "Your OTP for Utara Signup is: "+otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending OTP"})
		return
	}*/

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully. Please verify to complete signup.",
		"phone":   req.PhoneNumber,
	})
}

func VerifySignupOTP(c *gin.Context) {
	var req models.VerifySignupRequestOTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var otpEntry models.SignupOtpEntry
	err := config.DB.Collection("signup_otps").FindOne(context.Background(),
		bson.M{"phone_number": req.PhoneNumber}).Decode(&otpEntry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No signup request found"})
		return
	}

	if otpEntry.Otp != req.Otp || time.Now().After(otpEntry.OtpExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(otpEntry.Request.Password), bcrypt.DefaultCost)
	parts := strings.Split(otpEntry.Request.Email, "@")
	username := ""
	if len(parts) > 0 {
		username = parts[0]
	}

	user := models.User{
		Email:       otpEntry.Request.Email,
		UserName:    username,
		Gaam:        otpEntry.Request.Gaam,
		Password:    string(hashedPassword),
		Name:        otpEntry.Request.Name,
		Role:        otpEntry.Request.Role,
		PhoneNumber: otpEntry.Request.PhoneNumber,
		UserType:    "Neelkanth",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := config.DB.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	_, _ = config.DB.Collection("signup_otps").DeleteOne(context.Background(),
		bson.M{"_id": otpEntry.ID})

	user.Password = "" // hide password
	c.JSON(http.StatusOK, models.AuthResponse{
		Token: tokenString,
		User:  user,
	})
}

func CreateUser(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	// Validate role
	if req.Role != models.RoleSuperAdmin && req.Role != models.RoleStaff && req.Role != models.RoleUser {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Check if this is the first user being created
	count, err := config.DB.Collection("users").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user count"})
		return
	}

	// If this is not the first user, apply the regular role checks
	if count > 0 {
		if req.Role == models.RoleSuperAdmin || req.Role == models.RoleStaff {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin can create admin or staff users"})
				return
			}

			var currentUser models.User
			userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
			err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&currentUser)
			if err != nil || currentUser.Role != models.RoleSuperAdmin {
				c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin can create admin or staff users"})
				return
			}
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	parts := strings.Split(req.Email, "@")
	username := ""
	if len(parts) > 0 {
		username = parts[0]
		fmt.Println("Username:", username)
	} else {
		fmt.Println("Invalid email format")
	}
	// Create user
	user := models.User{
		Email:       req.Email,
		UserName:    username,
		Password:    string(hashedPassword),
		Name:        req.Name,
		Role:        req.Role,
		PhoneNumber: req.PhoneNumber,
		UserType:    "Neelkanth",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := config.DB.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	user.Password = "" // Remove password from response

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role, // Add role to JWT claims
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: tokenString,
		User:  user,
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role, // Add role to JWT claims
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	user.Password = "" // Remove password from response
	c.JSON(http.StatusOK, models.AuthResponse{
		Token: tokenString,
		User:  user,
	})
}

func UserLogin(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user with USER role only
	var user models.User
	filter := bson.M{
		"phone_number": req.PhoneNumber,
		"role":         models.RoleUser, // Only allow USER role
	}

	err := config.DB.Collection("users").FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or not a regular user"})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	user.Password = "" // Remove password from response
	c.JSON(http.StatusOK, models.AuthResponse{
		Token: tokenString,
		User:  user,
	})
}

func VerifyOTP(c *gin.Context) {
	var req models.VerifyOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	filter := bson.M{"phone_number": req.PhoneNumber}
	err := config.DB.Collection("users").FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number"})
		return
	}

	// Check OTP
	if user.Otp != req.Otp || time.Now().After(user.OtpExpiry) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Clear OTP
	_, _ = config.DB.Collection("users").UpdateOne(context.Background(), filter,
		bson.M{"$unset": bson.M{"otp": "", "otp_expiry": ""}})

	user.Password = ""
	c.JSON(http.StatusOK, models.AuthResponse{
		Token: tokenString,
		User:  user,
	})
}

func AssignModulesHandler(c *gin.Context) {

	var input struct {
		UserID string `json:"user_id"`
		//Role    string          `json:"role"`
		Modules map[string]bool `json:"modules"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Convert user_id to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if this user already has a document
	filter := bson.M{"user_id": userObjID}
	update := bson.M{
		"$set": bson.M{
			//"role":    input.Role,
			"modules": input.Modules,
			"updated": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = config.DB.Collection("user_module_access").UpdateOne(c.Request.Context(), filter, update, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Modules assigned successfully"})
}

func AssignUserType(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		UserType string `json:"user_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	allowedUserTypes := map[string]bool{
		"Shri Hari+": true,
		"Shri Hari":  true,
		"Sarju+":     true,
		"Sarju":      true,
		"Neelkanth+": true,
		"Neelkanth":  true,
	}

	if !allowedUserTypes[req.UserType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_type"})
		return
	}

	// Update user struct fields in "users" collection
	filter := bson.M{"_id": userObjID}
	update := bson.M{"$set": bson.M{
		"user_type":  req.UserType,
		"updated_at": time.Now(),
	}}

	result, err := config.DB.Collection("users").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user type"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User type updated successfully",
	})
}

func UpdateUsers(c *gin.Context) {
	id := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData := bson.M{}
	for key, value := range req {
		if value != nil && value != "" {
			// Hash password if updating
			if key == "password" {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(value.(string)), bcrypt.DefaultCost)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
					return
				}
				updateData["password"] = string(hashedPassword)
			} else {
				updateData[key] = value
			}
		}
	}

	updateData["updated_at"] = time.Now()

	result := config.DB.Collection("users").FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": updateData},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var updatedUser models.User
	if err := result.Decode(&updatedUser); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user"})
		return
	}

	updatedUser.Password = "" // hide password in response
	c.JSON(http.StatusOK, updatedUser)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result, err := config.DB.Collection("users").DeleteOne(context.Background(), bson.M{"_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"phone_number": req.PhoneNumber}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	otp := GenerateOTP()
	otpExpiry := time.Now().Add(60 * time.Second)

	_, err = config.DB.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"otp": otp, "otp_expiry": otpExpiry}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save OTP"})
		return
	}

	fmt.Println("OTP for user:", otp)
	// Here, integrate with your SMS service to send the OTP to user.PhoneNumber
	/*err = SendWhatsAppMessage(user.PhoneNumber, "Your OTP for Utara is: "+otp)
	if err != nil {
		print("error sending the OTP")
		return
	}*/

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully to your number"})
}

func ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := config.DB.Collection("users").FindOne(context.Background(), bson.M{"phone_number": req.PhoneNumber}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	_, err = config.DB.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"password":   string(hashedPassword),
				"updated_at": time.Now(),
			},
			"$unset": bson.M{"otp": "", "otp_expiry": ""},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
