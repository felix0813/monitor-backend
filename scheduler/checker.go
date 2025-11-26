package scheduler

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"monitor/db"
	"monitor/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	checkingEndpoints = make(map[string]bool)
	mutex             sync.Mutex
)

func StartChecker() {
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			var eps []models.Endpoint

			cursor, _ := db.DB().Collection("endpoints").Find(context.TODO(), bson.M{})
			cursor.All(context.TODO(), &eps)

			for _, ep := range eps {
				endpointID := ep.ID.Hex()

				mutex.Lock()
				if checkingEndpoints[endpointID] {
					mutex.Unlock()
					continue // 正在检查中，跳过
				}
				checkingEndpoints[endpointID] = true
				mutex.Unlock()

				go func(ep models.Endpoint) {
					defer func() {
						mutex.Lock()
						delete(checkingEndpoints, ep.ID.Hex())
						mutex.Unlock()
					}()
					checkEndpoint(ep)
				}(ep)
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
	status := "健康"
	if !result.Success {
		status = "异常"
	}
	// 更新 Endpoint 状态
	_, err := db.DB().Collection("endpoints").UpdateByID(ctx, ep.ID, bson.M{
		"$set": bson.M{
			"last_status":  status,
			"last_latency": result.LatencyMS,
			"updated_at":   time.Now(),
		},
	})

	if err != nil {
		// 记录错误日志
		log.Printf("Failed to update endpoint %s status: %v", ep.ID.Hex(), err)
		// 或者使用其他日志记录方式
	}
}
