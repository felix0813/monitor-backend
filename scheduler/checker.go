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

			now := time.Now()
			for _, ep := range eps {
				endpointID := ep.ID.Hex()

				// 检查是否到了该端点的检查时间
				if !shouldCheckEndpoint(ep, now) {
					continue
				}

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

// shouldCheckEndpoint 判断是否应该检查该端点
func shouldCheckEndpoint(ep models.Endpoint, now time.Time) bool {
	// 如果是第一次检查，应该检查
	if ep.UpdatedAt.IsZero() {
		return true
	}
	// 检查上次检查是否失败且间隔大于5秒
	if ep.LastStatus == "异常" && ep.Interval > 5 {
		// 失败且interval大于5秒时，5秒后即可检查
		nextRetryTime := ep.UpdatedAt.Add(5 * time.Second)
		return now.After(nextRetryTime)
	}
	// 根据 interval 判断是否到了检查时间
	// interval 是秒数
	nextCheckTime := ep.UpdatedAt.Add(time.Duration(ep.Interval) * time.Second)
	return now.After(nextCheckTime)
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
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return CheckResult{
			Success:   false,
			LatencyMS: latency,
		}, err
	}
	defer resp.Body.Close()
	success := resp.StatusCode < 400

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
	log.Println("Checking endpoint:", ep.URL)
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
