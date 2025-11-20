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

	// 获取环境变量配置
	mongoURI := os.Getenv("MONGO_URI")
	username := os.Getenv("MONGO_USERNAME")
	password := os.Getenv("MONGO_PASSWORD")

	// 构建连接选项
	var clientOptions *options.ClientOptions

	if mongoURI != "" {
		// 如果提供了完整的 URI，优先使用
		clientOptions = options.Client().ApplyURI(mongoURI)
	} else {
		// 否则构建基本连接配置
		host := os.Getenv("MONGO_HOST")
		if host == "" {
			host = "localhost:27017"
		}

		clientOptions = options.Client().ApplyURI("mongodb://" + host)

		// 如果提供了用户名密码，则添加认证
		if username != "" && password != "" {
			creds := options.Credential{
				Username: username,
				Password: password,
			}
			clientOptions.SetAuth(creds)
		}
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	Client = client

	// 检查 Time Series 集合
	db := client.Database("monitor")
	collections, _ := db.ListCollectionNames(ctx, bson.M{"name": "check_results"})
	if len(collections) == 0 {
		return errors.New("time series collection not found")
	}

	return nil
}

func DB() *mongo.Database {
	return Client.Database("health_check")
}
