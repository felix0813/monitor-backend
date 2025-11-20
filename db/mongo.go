package db

import (
	"context"
	"errors"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func InitMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return err
	}

	Client = client

	// 检查 Time Series 集合
	db := client.Database("monitor")
	collections, _ := db.ListCollectionNames(ctx, bson.M{"name": "check_results"})
	if len(collections) == 0 {
		return errors.New("time Series collection not found")
	}

	return nil
}

func DB() *mongo.Database {
	return Client.Database("health_check")
}
