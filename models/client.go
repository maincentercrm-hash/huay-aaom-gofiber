package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID        string             `bson:"user_id" json:"userId"`
	DisplayName   string             `bson:"display_name" json:"displayName"`
	PictureURL    string             `bson:"picture_url" json:"pictureUrl"`
	StatusMessage string             `bson:"status_message" json:"statusMessage"`
	PhoneNumber   string             `bson:"phone_number" json:"phoneNumber,omitempty"`
	CreatedAt     primitive.DateTime `bson:"created_at" json:"createdAt"`
	UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updatedAt"`
}
