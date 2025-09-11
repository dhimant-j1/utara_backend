package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MealType string

const (
	Breakfast MealType = "BREAKFAST"
	Lunch     MealType = "LUNCH"
	Dinner    MealType = "DINNER"
)

type FoodPass struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	MemberName string             `json:"member_name" bson:"member_name"`
	MealType   MealType           `json:"meal_type" bson:"meal_type"`
	Date       time.Time          `json:"date" bson:"date"`
	QRCode     string             `json:"qr_code" bson:"qr_code"`
	IsUsed     bool               `json:"is_used" bson:"is_used"`
	DiningHall string             `json:"dining_hall" bson:"dining_hall"`
	ColorCode  string             `json:"color_code" bson:"color_code"`
	DiningHall string             `json:"dining_hall" bson:"dining_hall"`
	ColorCode  string             `json:"color_code" bson:"color_code"`
	UsedAt     *time.Time         `json:"used_at,omitempty" bson:"used_at,omitempty"`
	CreatedBy  primitive.ObjectID `json:"created_by" bson:"created_by"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}

type GenerateFoodPassRequest struct {
	UserID      primitive.ObjectID `json:"user_id" binding:"required"`
	MemberNames []string           `json:"member_names" binding:"required"`
	StartDate   time.Time          `json:"start_date" binding:"required"`
	EndDate     time.Time          `json:"end_date" binding:"required"`
	DiningHall  string             `json:"dining_hall" binding:"required"`
	ColorCode   string             `json:"color_code" binding:"required"`
	DiningHall  string             `json:"dining_hall" binding:"required"`
	ColorCode   string             `json:"color_code" binding:"required"`
}

type ScanFoodPassRequest struct {
	PassID primitive.ObjectID `json:"pass_id" binding:"required"`
}

type FoodPassCategory struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BuildingName string             `json:"building_name" bson:"building_name"`
	ColorCode    string             `json:"color_code" bson:"color_code"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
}