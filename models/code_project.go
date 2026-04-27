package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CodeProject struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectName string             `bson:"project_name" json:"project_name"`
	CodeURL     string             `bson:"code_url" json:"code_url"`
	PipelineURL string             `bson:"pipeline_url" json:"pipeline_url"`
	DeployURL   string             `bson:"deploy_url" json:"deploy_url"`
	DataURL     string             `bson:"data_url" json:"data_url"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
