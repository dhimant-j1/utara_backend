package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"utara_backend/config"
	"utara_backend/models"
)

func Signup(c *gin.Context) {
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

	// Create user
	user := models.User{
		Email:       req.Email,
		Password:    string(hashedPassword),
		Name:        req.Name,
		Role:        req.Role,
		PhoneNumber: req.PhoneNumber,
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

	// Create user
	user := models.User{
		Email:       req.Email,
		Password:    string(hashedPassword),
		Name:        req.Name,
		Role:        req.Role,
		PhoneNumber: req.PhoneNumber,
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
