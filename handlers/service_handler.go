package handlers

import (
	"context"
	"monitor/db"
	"monitor/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServiceHandler struct{}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{}
}

func (h *ServiceHandler) CreateService(c *gin.Context) {
	var s models.Service
	if err := c.BindJSON(&s); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	s.ID = primitive.NewObjectID()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	_, err := db.DB().Collection("services").InsertOne(context.TODO(), s)
	if err != nil {
		c.JSON(500, gin.H{"error": "insert failed"})
		return
	}

	c.JSON(200, s)
}

func (h *ServiceHandler) ListServices(c *gin.Context) {
	var result []models.Service

	cur, _ := db.DB().Collection("services").Find(context.TODO(), bson.M{})
	cur.All(context.TODO(), &result)

	c.JSON(200, result)
}

func (h *ServiceHandler) GetService(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var s models.Service
	err := db.DB().Collection("services").
		FindOne(context.TODO(), bson.M{"_id": id}).Decode(&s)

	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, s)
}

func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var body bson.M
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	body["updated_at"] = time.Now()

	_, err := db.DB().Collection("services").UpdateByID(context.TODO(), id, bson.M{
		"$set": body,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}

	c.JSON(200, gin.H{"message": "ok"})
}

func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	db.DB().Collection("services").DeleteOne(context.TODO(), bson.M{"_id": id})
	db.DB().Collection("endpoints").DeleteMany(context.TODO(), bson.M{"service_id": id})

	c.JSON(200, gin.H{"message": "deleted"})
}
