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
	UserName    string             `json:"user_name" bson:"user_name"`
	Email       string             `json:"email" bson:"email"`
	Password    string             `json:"-" bson:"password"`
	Name        string             `json:"name" bson:"name"`
	Role        UserRole           `json:"role" bson:"role"`
	IsImportant bool               `json:"is_important" bson:"is_important"`
	PhoneNumber string             `json:"phone_number" bson:"phone_number"`
	UserType    string             `json:"user_type" bson:"user_type"`
	Otp         string             `json:"otp,omitempty" bson:"otp,omitempty"`
	OtpExpiry   time.Time          `json:"otp_expiry,omitempty" bson:"otp_expiry,omitempty"`
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

type UserModuleAccess struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Role      string             `bson:"role" json:"role"`
	Modules   map[string]bool    `bson:"modules" json:"modules"`
	CreatedAt time.Time          `bson:"created" json:"created"`
	UpdatedAt time.Time          `bson:"updated" json:"updated"`
}

type UserLoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
}

type VerifyOtpRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Otp         string `json:"otp" binding:"required,len=6"`
}

type ForgotPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type ResetPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Otp         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type SignupOtpEntry struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	PhoneNumber string             `bson:"phone_number"`
	Request     SignupRequest      `bson:"request"`
	Otp         string             `bson:"otp"`
	OtpExpiry   time.Time          `bson:"otp_expiry"`
	CreatedAt   time.Time          `bson:"created_at"`
}

type VerifySignupRequestOTP struct {
	PhoneNumber string `json:"phone_number"`
	Otp         string `json:"otp"`
}
