package handlers

import (
	"context"
	"time"

	"monitor/db"
	"monitor/models"
	"monitor/scheduler"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EndpointHandler struct{}

func NewEndpointHandler() *EndpointHandler {
	return &EndpointHandler{}
}

func (h *EndpointHandler) CreateEndpoint(c *gin.Context) {
	serviceID, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var ep models.Endpoint
	if err := c.BindJSON(&ep); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	ep.ID = primitive.NewObjectID()
	ep.ServiceID = serviceID
	ep.CreatedAt = time.Now()
	ep.UpdatedAt = time.Now()

	_, err := db.DB().Collection("endpoints").InsertOne(context.TODO(), ep)
	if err != nil {
		c.JSON(500, gin.H{"error": "insert failed"})
		return
	}

	c.JSON(200, ep)
}

func (h *EndpointHandler) ListEndpoints(c *gin.Context) {
	serviceID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid service ID format", "details": err.Error()})
		return
	}

	var result []models.Endpoint

	filter := bson.M{"service_id": serviceID}
	cur, err := db.DB().Collection("endpoints").Find(context.TODO(), filter)
	if err != nil {
		c.JSON(500, gin.H{"error": "database query failed", "details": err.Error()})
		return
	}
	defer cur.Close(context.TODO())

	err = cur.All(context.TODO(), &result)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to decode results", "details": err.Error()})
		return
	}

	c.JSON(200, result)
}

func (h *EndpointHandler) GetEndpoint(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var ep models.Endpoint
	err := db.DB().Collection("endpoints").
		FindOne(context.TODO(), bson.M{"_id": id}).Decode(&ep)

	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, ep)
}

func (h *EndpointHandler) UpdateEndpoint(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var body bson.M
	c.BindJSON(&body)
	body["updated_at"] = time.Now()

	db.DB().Collection("endpoints").UpdateByID(context.TODO(), id, bson.M{
		"$set": body,
	})

	c.JSON(200, gin.H{"message": "ok"})
}

func (h *EndpointHandler) DeleteEndpoint(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	db.DB().Collection("endpoints").DeleteOne(context.TODO(), bson.M{"_id": id})

	c.JSON(200, gin.H{"message": "deleted"})
}

func (h *EndpointHandler) CheckEndpointNow(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var ep models.Endpoint
	err := db.DB().Collection("endpoints").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&ep)

	if err != nil {
		c.JSON(404, gin.H{"error": "endpoint not found"})
		return
	}

	ctx := context.Background()
	result, err := scheduler.PerformCheck(ctx, db.DB(), ep.ID.Hex(), ep.URL)
	if err != nil {
		c.JSON(500, gin.H{"error": "check failed"})
		return
	}
	// 将布尔值转换为字符串状态
	status := "健康"
	if !result.Success {
		status = "异常"
	}
	db.DB().Collection("endpoints").UpdateByID(ctx, ep.ID, bson.M{
		"$set": bson.M{
			"last_status":  status,
			"last_latency": result.LatencyMS,
			"updated_at":   time.Now(),
		},
	})

	c.JSON(200, gin.H{
		"success": result.Success,
		"latency": result.LatencyMS,
	})
}
