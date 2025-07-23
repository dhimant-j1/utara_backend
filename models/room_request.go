package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestStatus string

const (
	StatusPending  RequestStatus = "PENDING"
	StatusApproved RequestStatus = "APPROVED"
	StatusRejected RequestStatus = "REJECTED"
)

type RoomRequest struct {
	ID              primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	UserID          primitive.ObjectID  `json:"user_id" bson:"user_id"`
	Name            string              `json:"name" bson:"name"`
	CheckInDate     time.Time           `json:"check_in_date" bson:"check_in_date"`
	CheckOutDate    time.Time           `json:"check_out_date" bson:"check_out_date"`
	NumberOfPeople  int                 `json:"number_of_people" bson:"number_of_people"`
	PreferredType   RoomType            `json:"preferred_type" bson:"preferred_type"`
	SpecialRequests string              `json:"special_requests" bson:"special_requests"`
	Status          RequestStatus       `json:"status" bson:"status"`
	ProcessedBy     *primitive.ObjectID `json:"processed_by,omitempty" bson:"processed_by,omitempty"`
	ProcessedAt     *time.Time          `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" bson:"updated_at"`
}

type RoomAssignment struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RoomID       primitive.ObjectID `json:"room_id" bson:"room_id"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	RequestID    primitive.ObjectID `json:"request_id" bson:"request_id"`
	CheckInDate  time.Time          `json:"check_in_date" bson:"check_in_date"`
	CheckOutDate time.Time          `json:"check_out_date" bson:"check_out_date"`
	AssignedBy   primitive.ObjectID `json:"assigned_by" bson:"assigned_by"`
	AssignedAt   time.Time          `json:"assigned_at" bson:"assigned_at"`
	CheckedIn    bool               `json:"checked_in" bson:"checked_in"`
	CheckedInAt  *time.Time         `json:"checked_in_at,omitempty" bson:"checked_in_at,omitempty"`
	CheckedOut   bool               `json:"checked_out" bson:"checked_out"`
	CheckedOutAt *time.Time         `json:"checked_out_at,omitempty" bson:"checked_out_at,omitempty"`
}

type CreateRoomRequestRequest struct {
	CheckInDate     time.Time `json:"check_in_date" binding:"required"`
	CheckOutDate    time.Time `json:"check_out_date" binding:"required"`
	NumberOfPeople  int       `json:"number_of_people" binding:"required"`
	PreferredType   RoomType  `json:"preferred_type" binding:"required"`
	SpecialRequests string    `json:"special_requests"`
}

type ProcessRoomRequestRequest struct {
	Status RequestStatus       `json:"status" binding:"required"`
	RoomID *primitive.ObjectID `json:"room_id"`
}

type RoomAssignmentRequest struct {
	RoomID       primitive.ObjectID `json:"room_id" binding:"required"`
	UserID       primitive.ObjectID `json:"user_id" binding:"required"`
	RequestID    primitive.ObjectID `json:"request_id" binding:"required"`
	CheckInDate  time.Time          `json:"check_in_date" binding:"required"`
	CheckOutDate time.Time          `json:"check_out_date" binding:"required"`
}
