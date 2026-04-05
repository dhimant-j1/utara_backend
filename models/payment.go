package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCreated   PaymentStatus = "CREATED"
	PaymentStatusPaid      PaymentStatus = "PAID"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

type PaymentType string

const (
	PaymentTypeRoomBooking PaymentType = "ROOM_BOOKING"
	PaymentTypeFoodPass    PaymentType = "FOOD_PASS"
	PaymentTypeDeposit     PaymentType = "DEPOSIT"
	PaymentTypeOther       PaymentType = "OTHER"
)

// Payment represents a payment transaction
type Payment struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
	RequestID       *primitive.ObjectID `json:"request_id,omitempty" bson:"request_id,omitempty"`
	Amount          int                `json:"amount" bson:"amount"` // Amount in paise
	Currency        string             `json:"currency" bson:"currency"`
	Type            PaymentType        `json:"type" bson:"type"`
	Status          PaymentStatus      `json:"status" bson:"status"`
	RazorpayOrderID string             `json:"razorpay_order_id,omitempty" bson:"razorpay_order_id,omitempty"`
	RazorpayPaymentID string           `json:"razorpay_payment_id,omitempty" bson:"razorpay_payment_id,omitempty"`
	RazorpaySignature string           `json:"razorpay_signature,omitempty" bson:"razorpay_signature,omitempty"`
	Description     string             `json:"description" bson:"description"`
	Notes           map[string]string  `json:"notes,omitempty" bson:"notes,omitempty"`
	FailureReason   string             `json:"failure_reason,omitempty" bson:"failure_reason,omitempty"`
	RefundID        string             `json:"refund_id,omitempty" bson:"refund_id,omitempty"`
	RefundedAmount  int                `json:"refunded_amount,omitempty" bson:"refunded_amount,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
	PaidAt          *time.Time         `json:"paid_at,omitempty" bson:"paid_at,omitempty"`
}

// CreatePaymentRequest represents a request to create a new payment
type CreatePaymentRequest struct {
	Amount      int               `json:"amount" binding:"required"` // Amount in paise
	Currency    string            `json:"currency" binding:"required"`
	Type        PaymentType       `json:"type" binding:"required"`
	RequestID   *string           `json:"request_id,omitempty"`
	Description string            `json:"description" binding:"required"`
	Notes       map[string]string `json:"notes,omitempty"`
}

// VerifyPaymentRequest represents a request to verify a payment
type VerifyPaymentRequest struct {
	OrderID   string `json:"order_id" binding:"required"`
	PaymentID string `json:"payment_id" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// PaymentResponse represents a payment response with order details
type PaymentResponse struct {
	Payment     Payment `json:"payment"`
	RazorpayKey string  `json:"razorpay_key"`
}

// RazorpayOrderResponse represents Razorpay order creation response
type RazorpayOrderResponse struct {
	ID         string            `json:"id"`
	Entity     string            `json:"entity"`
	Amount     int               `json:"amount"`
	Currency   string            `json:"currency"`
	Status     string            `json:"status"`
	Notes      map[string]string `json:"notes"`
	CreatedAt  int64             `json:"created_at"`
}
