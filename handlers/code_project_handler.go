package handlers

import (
	"errors"
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

const codeProjectCollection = "code_projects"

type CodeProjectHandler struct{}

type createCodeProjectRequest struct {
	ProjectName string `json:"project_name"`
	CodeURL     string `json:"code_url"`
	PipelineURL string `json:"pipeline_url"`
	DeployURL   string `json:"deploy_url"`
	DataURL     string `json:"data_url"`
}

type updateCodeProjectRequest struct {
	ProjectName *string `json:"project_name"`
	CodeURL     *string `json:"code_url"`
	PipelineURL *string `json:"pipeline_url"`
	DeployURL   *string `json:"deploy_url"`
	DataURL     *string `json:"data_url"`
}

func NewCodeProjectHandler() *CodeProjectHandler {
	return &CodeProjectHandler{}
}

func (h *CodeProjectHandler) ListCodeProjects(c *gin.Context) {
	ctx := c.Request.Context()
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cur, err := db.DB().Collection(codeProjectCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list code projects"})
		return
	}
	defer cur.Close(ctx)

	result := make([]models.CodeProject, 0)
	if err := cur.All(ctx, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode code projects"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CodeProjectHandler) CreateCodeProject(c *gin.Context) {
	var req createCodeProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	project := models.CodeProject{
		ID:          primitive.NewObjectID(),
		ProjectName: strings.TrimSpace(req.ProjectName),
		CodeURL:     strings.TrimSpace(req.CodeURL),
		PipelineURL: strings.TrimSpace(req.PipelineURL),
		DeployURL:   strings.TrimSpace(req.DeployURL),
		DataURL:     strings.TrimSpace(req.DataURL),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if project.ProjectName == "" || project.CodeURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_name and code_url are required"})
		return
	}

	if _, err := db.DB().Collection(codeProjectCollection).InsertOne(c.Request.Context(), project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create code project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *CodeProjectHandler) GetCodeProject(c *gin.Context) {
	id, ok := parseCodeProjectID(c)
	if !ok {
		return
	}

	project, ok := h.findCodeProject(c, id)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *CodeProjectHandler) UpdateCodeProject(c *gin.Context) {
	id, ok := parseCodeProjectID(c)
	if !ok {
		return
	}

	var req updateCodeProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if req.ProjectName != nil {
		set["project_name"] = strings.TrimSpace(*req.ProjectName)
	}
	if req.CodeURL != nil {
		set["code_url"] = strings.TrimSpace(*req.CodeURL)
	}
	if req.PipelineURL != nil {
		set["pipeline_url"] = strings.TrimSpace(*req.PipelineURL)
	}
	if req.DeployURL != nil {
		set["deploy_url"] = strings.TrimSpace(*req.DeployURL)
	}
	if req.DataURL != nil {
		set["data_url"] = strings.TrimSpace(*req.DataURL)
	}

	if name, exists := set["project_name"]; exists && name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_name is required"})
		return
	}
	if codeURL, exists := set["code_url"]; exists && codeURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_url is required"})
		return
	}

	ctx := c.Request.Context()
	result, err := db.DB().Collection(codeProjectCollection).UpdateByID(ctx, id, bson.M{"$set": set})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update code project"})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "code project not found"})
		return
	}

	project, ok := h.findCodeProject(c, id)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *CodeProjectHandler) DeleteCodeProject(c *gin.Context) {
	id, ok := parseCodeProjectID(c)
	if !ok {
		return
	}

	result, err := db.DB().Collection(codeProjectCollection).DeleteOne(c.Request.Context(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete code project"})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "code project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func parseCodeProjectID(c *gin.Context) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code project id"})
		return primitive.NilObjectID, false
	}
	return id, true
}

func (h *CodeProjectHandler) findCodeProject(c *gin.Context, id primitive.ObjectID) (models.CodeProject, bool) {
	var project models.CodeProject
	err := db.DB().Collection(codeProjectCollection).
		FindOne(c.Request.Context(), bson.M{"_id": id}).
		Decode(&project)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"error": "code project not found"})
		return models.CodeProject{}, false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load code project"})
		return models.CodeProject{}, false
	}

	return project, true
}
