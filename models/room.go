package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoomType string

const (
	ShreeHariPlus RoomType = "SHREEHARIPLUS"
	ShreeHari     RoomType = "SHREEHARI"
	SarjuPlus     RoomType = "SARJUPLUS"
	Sarju         RoomType = "SARJU"
	NeelkanthPlus RoomType = "NEELKANTHPLUS"
	Neelkanth     RoomType = "NEELKANTH"
)

type BedType string

const (
	Single   BedType = "SINGLE"
	Double   BedType = "DOUBLE"
	ExtraBed BedType = "EXTRA_BED"
)

type Bed struct {
	Type     BedType `json:"type" bson:"type"`
	Quantity int     `json:"quantity" bson:"quantity"`
}

type RoomImage struct {
	URL         string    `json:"url" bson:"url"`
	Description string    `json:"description" bson:"description"`
	UploadedAt  time.Time `json:"uploaded_at" bson:"uploaded_at"`
}

type Room struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RoomNumber      string             `json:"room_number" bson:"room_number"`
	Floor           int                `json:"floor" bson:"floor"`
	Type            RoomType           `json:"type" bson:"type"`
	Beds            []Bed              `json:"beds" bson:"beds"`
	HasGeyser       bool               `json:"has_geyser" bson:"has_geyser"`
	HasAC           bool               `json:"has_ac" bson:"has_ac"`
	HasSofaSet      bool               `json:"has_sofa_set" bson:"has_sofa_set"`
	SofaSetQuantity int                `json:"sofa_set_quantity,omitempty" bson:"sofa_set_quantity,omitempty"`
	ExtraAmenities  string             `json:"extra_amenities" bson:"extra_amenities"`
	IsVisible       bool               `json:"is_visible" bson:"is_visible"`
	IsOccupied      bool               `json:"is_occupied" bson:"is_occupied"`
	NeedsCleaning   bool               `json:"needs_cleaning" bson:"needs_cleaning"`
	Building        string             `bson:"building" json:"building"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
	RoomCategoryId  string             `json:"room_category_id,omitempty" bson:"room_category_id,omitempty"`
}

type CreateRoomRequest struct {
	RoomNumber      string   `json:"room_number" binding:"required"`
	Floor           int      `json:"floor" binding:"required"`
	Type            RoomType `json:"type" binding:"required"`
	Beds            []Bed    `json:"beds" binding:"required"`
	HasGeyser       bool     `json:"has_geyser"`
	HasAC           bool     `json:"has_ac"`
	HasSofaSet      bool     `json:"has_sofa_set"`
	SofaSetQuantity int      `json:"sofa_set_quantity"`
	RoomCategoryId  string   `json:"room_category_id,omitempty"`
	ExtraAmenities  string   `json:"extra_amenities"`
	IsVisible       bool     `json:"is_visible"`
	Building        string   `json:"building"`
}

type UpdateRoomRequest struct {
	RoomNumber      *string   `json:"room_number"`
	Floor           *int      `json:"floor"`
	Type            *RoomType `json:"type"`
	Beds            []Bed     `json:"beds"`
	HasGeyser       *bool     `json:"has_geyser"`
	HasAC           *bool     `json:"has_ac"`
	HasSofaSet      *bool     `json:"has_sofa_set"`
	SofaSetQuantity *int      `json:"sofa_set_quantity"`
	ExtraAmenities  *string   `json:"extra_amenities"`
	IsVisible       *bool     `json:"is_visible"`
	NeedsCleaning   *bool     `json:"needs_cleaning"`
}

type RoomCategory struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RoomName  string             `json:"room_name" bson:"room_name"`
	Images    []RoomImage        `json:"images"`
	Price     string             `json:"price" bson:"price"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// RoomTypeCost stores the deposit cost for each room type (Sarju, Shri Hari, etc.)
type RoomTypeCost struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RoomType      RoomType           `json:"room_type" bson:"room_type"`           // SARJU, SHREEHARI, etc.
	DepositAmount int                `json:"deposit_amount" bson:"deposit_amount"` // Amount in paise
	Currency      string             `json:"currency" bson:"currency"`
	Description   string             `json:"description" bson:"description"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

// GetDefaultRoomTypeCosts returns default deposit amounts for each room type
func GetDefaultRoomTypeCosts() []RoomTypeCost {
	now := time.Now()
	return []RoomTypeCost{
		{RoomType: ShreeHariPlus, DepositAmount: 50000, Currency: "INR", Description: "Shree Hari Plus rooms", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{RoomType: ShreeHari, DepositAmount: 30000, Currency: "INR", Description: "Shree Hari rooms", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{RoomType: SarjuPlus, DepositAmount: 50000, Currency: "INR", Description: "Sarju Plus rooms", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{RoomType: Sarju, DepositAmount: 30000, Currency: "INR", Description: "Sarju rooms", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{RoomType: NeelkanthPlus, DepositAmount: 50000, Currency: "INR", Description: "Neelkanth Plus rooms", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{RoomType: Neelkanth, DepositAmount: 30000, Currency: "INR", Description: "Neelkanth rooms", IsActive: true, CreatedAt: now, UpdatedAt: now},
	}
}
