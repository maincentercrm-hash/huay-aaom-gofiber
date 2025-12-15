package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Log struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID        string             `bson:"user_id" json:"user_id"`
	MissionID     string             `bson:"mission_id" json:"mission_id"`
	MissionDetail string             `bson:"mission_detail" json:"mission_detail"`
	Reward        float64            `bson:"reward" json:"reward"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	CallbackTime  time.Time          `bson:"callback_time,omitempty" json:"callback_time,omitempty"`
	Status        string             `bson:"status" json:"status"` // "pending", "approve", "reject"
}
