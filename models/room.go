package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoomType string

const (
	Standard   RoomType = "STANDARD"
	Deluxe     RoomType = "DELUXE"
	Suite      RoomType = "SUITE"
	FamilyRoom RoomType = "FAMILY_ROOM"
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
	Images          []RoomImage        `json:"images" bson:"images"`
	IsOccupied      bool               `json:"is_occupied" bson:"is_occupied"`
	Building        string             `bson:"building" json:"building"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

type CreateRoomRequest struct {
	RoomNumber      string      `json:"room_number" binding:"required"`
	Floor           int         `json:"floor" binding:"required"`
	Type            RoomType    `json:"type" binding:"required"`
	Beds            []Bed       `json:"beds" binding:"required"`
	HasGeyser       bool        `json:"has_geyser"`
	HasAC           bool        `json:"has_ac"`
	HasSofaSet      bool        `json:"has_sofa_set"`
	SofaSetQuantity int         `json:"sofa_set_quantity"`
	ExtraAmenities  string      `json:"extra_amenities"`
	IsVisible       bool        `json:"is_visible"`
	Images          []RoomImage `json:"images"`
}

type UpdateRoomRequest struct {
	RoomNumber      *string     `json:"room_number"`
	Floor           *int        `json:"floor"`
	Type            *RoomType   `json:"type"`
	Beds            []Bed       `json:"beds"`
	HasGeyser       *bool       `json:"has_geyser"`
	HasAC           *bool       `json:"has_ac"`
	HasSofaSet      *bool       `json:"has_sofa_set"`
	SofaSetQuantity *int        `json:"sofa_set_quantity"`
	ExtraAmenities  *string     `json:"extra_amenities"`
	IsVisible       *bool       `json:"is_visible"`
	Images          []RoomImage `json:"images"`
}
