package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserBet struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID     string             `bson:"user_id" json:"user_id"`
	CurrentBet float64            `bson:"current_bet" json:"current_bet"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
