package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Mission struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID           string             `bson:"user_id" json:"user_id"`
	PhoneNumber      string             `bson:"phone_number" json:"phone_number"`
	Status           string             `bson:"status" json:"status"` // "processing", "completed", "failed", "pending"
	CurrentTier      int                `bson:"current_tier" json:"current_tier"`
	Tiers            []Tier             `bson:"tiers" json:"tiers"`
	ConsecutiveFails int                `bson:"consecutive_fails" json:"consecutive_fails"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type Tier struct {
	Name         string    `bson:"name" json:"name"`
	Reward       int       `bson:"reward" json:"reward"`
	Target       int       `bson:"target" json:"target"`
	Status       string    `bson:"status" json:"status"`
	CurrentLevel int       `bson:"current_level" json:"current_level"`
	MaxLevel     int       `bson:"max_level" json:"max_level"`
	Levels       []Level   `bson:"levels" json:"levels"`
	ExpireReward time.Time `bson:"expire_reward" json:"expire_reward"`
}

type Level struct {
	Name         string    `bson:"name" json:"name"`
	StartDate    time.Time `bson:"start_date" json:"start_date"`
	ExpireDate   time.Time `bson:"expire_date" json:"expire_date"`
	FollowUpDate time.Time `bson:"follow_up_date" json:"follow_up_date"`
	Status       string    `bson:"status" json:"status"`
	CurrentBet   float64   `bson:"current_bet" json:"current_bet"` // New field
}
