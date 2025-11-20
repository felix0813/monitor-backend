package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Endpoint struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceID      primitive.ObjectID `bson:"service_id" json:"service_id"`
	Name           string             `bson:"name" json:"name"`
	URL            string             `bson:"url" json:"url"`
	Method         string             `bson:"method" json:"method"`
	Interval       int                `bson:"interval" json:"interval"`
	Timeout        int                `bson:"timeout" json:"timeout"`
	ExpectedStatus int                `bson:"expected_status" json:"expected_status"`

	LastStatus  string    `bson:"last_status" json:"last_status"`
	LastLatency int       `bson:"last_latency" json:"last_latency"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}
