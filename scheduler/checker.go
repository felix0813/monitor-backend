package scheduler

import (
	"context"
	"net/http"
	"time"

	"monitor/db"
	"monitor/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartChecker() {
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			var eps []models.Endpoint

			cursor, _ := db.DB().Collection("endpoints").Find(context.TODO(), bson.M{})
			cursor.All(context.TODO(), &eps)

			for _, ep := range eps {
				// 达到检查间隔就执行一次
				if time.Since(ep.UpdatedAt).Seconds() >= float64(ep.Interval) {
					go checkEndpoint(ep)
				}
			}
		}
	}()
}

// ---------------------------
// 公共结构
// ---------------------------

type CheckResult struct {
	Success   bool
	LatencyMS int64
}

// ---------------------------
// PerformCheck（对外暴露）
// ---------------------------

func PerformCheck(ctx context.Context, db *mongo.Database, endpointID, url string) (CheckResult, error) {
	start := time.Now()

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	success := err == nil && resp.StatusCode < 400
	latency := time.Since(start).Milliseconds()

	result := CheckResult{
		Success:   success,
		LatencyMS: latency,
	}

	// 插入时序数据
	_, err = db.Collection("check_results").InsertOne(ctx, bson.M{
		"endpoint_id": endpointID,
		"success":     success,
		"latency_ms":  latency,
		"checked_at":  time.Now(),
	})

	return result, err
}

// ---------------------------
// checkEndpoint（内部使用）
// ---------------------------

func checkEndpoint(ep models.Endpoint) {
	ctx := context.Background()

	result, _ := PerformCheck(ctx, db.DB(), ep.ID.Hex(), ep.URL)

	// 更新 Endpoint 状态
	db.DB().Collection("endpoints").UpdateByID(ctx, ep.ID, bson.M{
		"$set": bson.M{
			"last_status":  result.Success,
			"last_latency": result.LatencyMS,
			"updated_at":   time.Now(),
		},
	})
}
