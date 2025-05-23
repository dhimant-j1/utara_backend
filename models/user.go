package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	RoleSuperAdmin UserRole = "SUPER_ADMIN"
	RoleStaff      UserRole = "STAFF"
	RoleUser       UserRole = "USER"
)

type User struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email       string             `json:"email" bson:"email"`
	Password    string             `json:"-" bson:"password"`
	Name        string             `json:"name" bson:"name"`
	Role        UserRole           `json:"role" bson:"role"`
	IsImportant bool               `json:"is_important" bson:"is_important"`
	PhoneNumber string             `json:"phone_number" bson:"phone_number"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type SignupRequest struct {
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=6"`
	Name        string   `json:"name" binding:"required"`
	PhoneNumber string   `json:"phone_number" binding:"required"`
	Role        UserRole `json:"role" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateUserRequest struct {
	Name        *string   `json:"name"`
	PhoneNumber *string   `json:"phone_number"`
	Role        *UserRole `json:"role"`
	IsImportant *bool     `json:"is_important"`
}
