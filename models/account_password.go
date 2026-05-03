package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountPassword struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Account     string             `bson:"account" json:"account"`
	Password    string             `bson:"password" json:"password"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
