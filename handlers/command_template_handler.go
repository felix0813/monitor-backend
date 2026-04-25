package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"monitor/db"
	"monitor/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommandTemplateHandler struct{}

type createCommandTemplateRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type updateCommandTemplateRequest struct {
	Name    *string `json:"name"`
	Content *string `json:"content"`
}

func NewCommandTemplateHandler() *CommandTemplateHandler {
	return &CommandTemplateHandler{}
}

func (h *CommandTemplateHandler) CreateCommandTemplate(c *gin.Context) {
	var req createCommandTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	name := strings.TrimSpace(req.Name)
	content := strings.TrimSpace(req.Content)
	if name == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and content are required"})
		return
	}

	now := time.Now()
	template := models.CommandTemplate{
		ID:        primitive.NewObjectID(),
		Name:      name,
		Content:   content,
		Variables: models.ExtractCommandTemplateVariables(content),
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := db.DB().Collection("command_templates").InsertOne(context.TODO(), template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *CommandTemplateHandler) ListCommandTemplates(c *gin.Context) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cur, err := db.DB().Collection("command_templates").Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer cur.Close(context.TODO())

	result := make([]models.CommandTemplate, 0)
	if err := cur.All(context.TODO(), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode failed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CommandTemplateHandler) GetCommandTemplate(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var template models.CommandTemplate
	err = db.DB().Collection("command_templates").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&template)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *CommandTemplateHandler) UpdateCommandTemplate(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateCommandTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var template models.CommandTemplate
	err = db.DB().Collection("command_templates").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&template)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	if req.Name != nil {
		template.Name = strings.TrimSpace(*req.Name)
	}
	if req.Content != nil {
		template.Content = strings.TrimSpace(*req.Content)
		template.Variables = models.ExtractCommandTemplateVariables(template.Content)
	}
	if template.Name == "" || template.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and content are required"})
		return
	}

	template.UpdatedAt = time.Now()

	_, err = db.DB().Collection("command_templates").UpdateByID(context.TODO(), id, bson.M{
		"$set": bson.M{
			"name":       template.Name,
			"content":    template.Content,
			"variables":  template.Variables,
			"updated_at": template.UpdatedAt,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *CommandTemplateHandler) DeleteCommandTemplate(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := db.DB().Collection("command_templates").DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
