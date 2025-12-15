package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExpirationEvent struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MissionID  primitive.ObjectID `bson:"mission_id" json:"mission_id"`
	TierIndex  int                `bson:"tier_index" json:"tier_index"`
	LevelIndex int                `bson:"level_index" json:"level_index"`
	ExpireTime time.Time          `bson:"expire_time" json:"expire_time"`
	Status     string             `bson:"status" json:"status"` // "pending" or "processed"
	Type       string             `bson:"type" json:"type"`     // "level_expiration", "follow_up", or "reward_expiration"
}
