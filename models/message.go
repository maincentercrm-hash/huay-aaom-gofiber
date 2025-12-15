package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Status      string             `bson:"status" json:"status"` // "sent", "read", "unread" etc.
	Tier        string             `bson:"tier" json:"tier"`
	Level       string             `bson:"level" json:"level"`
	MissionID   primitive.ObjectID `bson:"mission_id" json:"mission_id"`
	SentAt      time.Time          `bson:"sent_at" json:"sent_at"`
	ReadAt      time.Time          `bson:"read_at,omitempty" json:"read_at,omitempty"`
	FlexContent FlexContent        `bson:"flex_content" json:"flex_content"`
}

type FlexContent struct {
	Title          string `bson:"title" json:"title"`
	Description    string `bson:"description" json:"description"`
	SubDescription string `bson:"sub_description,omitempty" json:"sub_description,omitempty"`
}
