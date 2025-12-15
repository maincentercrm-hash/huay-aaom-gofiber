package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email      string             `bson:"email" json:"email"`
	Password   string             `bson:"password" json:"-"` // ไม่ส่งกลับในการ response JSON
	Role       string             `bson:"role" json:"role"`
	Status     string             `bson:"status" json:"status"`
	CreateDate time.Time          `bson:"createDate" json:"createDate"`
}
