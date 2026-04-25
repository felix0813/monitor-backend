package handlers

import (
	"net/http"
	"strings"
	"time"

	"monitor/db"
	"monitor/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TodoHandler struct{}

type createTodoProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateTodoProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type createTodoItemRequest struct {
	Description string             `json:"description"`
	Status      *models.TodoStatus `json:"status"`
}

type updateTodoItemRequest struct {
	Description *string            `json:"description"`
	Status      *models.TodoStatus `json:"status"`
}

func NewTodoHandler() *TodoHandler {
	return &TodoHandler{}
}

func (h *TodoHandler) ListTodoProjects(c *gin.Context) {
	store, err := db.LoadTodoStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load todo projects"})
		return
	}

	result := make([]*models.TodoProject, 0, len(store.OrderedIDs))
	for _, id := range store.OrderedIDs {
		project, exists := store.Projects[id]
		if !exists {
			continue
		}
		result = append(result, project)
	}

	c.JSON(http.StatusOK, result)
}

func (h *TodoHandler) CreateTodoProject(c *gin.Context) {
	var req createTodoProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	store, err := db.LoadTodoStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load todo projects"})
		return
	}

	now := time.Now()
	project := &models.TodoProject{
		ID:          primitive.NewObjectID().Hex(),
		Name:        req.Name,
		Description: req.Description,
		OrderedIDs:  make([]string, 0),
		Items:       make(map[string]*models.TodoItem),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	store.Projects[project.ID] = project
	store.OrderedIDs = append(store.OrderedIDs, project.ID)

	if err := db.SaveTodoStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save todo project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *TodoHandler) UpdateTodoProject(c *gin.Context) {
	var req updateTodoProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	store, err := db.LoadTodoStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load todo projects"})
		return
	}

	project, exists := store.Projects[c.Param("project_id")]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo project not found"})
		return
	}

	if req.Name != nil {
		project.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		project.Description = strings.TrimSpace(*req.Description)
	}
	if project.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	project.UpdatedAt = time.Now()
	if err := db.SaveTodoStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save todo project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *TodoHandler) DeleteTodoProject(c *gin.Context) {
	store, err := db.LoadTodoStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load todo projects"})
		return
	}

	id := c.Param("project_id")
	if _, exists := store.Projects[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo project not found"})
		return
	}

	delete(store.Projects, id)
	store.OrderedIDs = removeID(store.OrderedIDs, id)

	if err := db.SaveTodoStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save todo project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *TodoHandler) ListTodoItems(c *gin.Context) {
	project, ok := h.loadTodoProject(c)
	if !ok {
		return
	}

	result := make([]*models.TodoItem, 0, len(project.OrderedIDs))
	for _, id := range project.OrderedIDs {
		item, exists := project.Items[id]
		if !exists {
			continue
		}
		result = append(result, item)
	}

	c.JSON(http.StatusOK, result)
}

func (h *TodoHandler) CreateTodoItem(c *gin.Context) {
	var req createTodoItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	req.Description = strings.TrimSpace(req.Description)
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	status := models.TodoStatusPending
	if req.Status != nil {
		status = *req.Status
	}
	if !isValidTodoStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	store, project, ok := h.loadTodoProjectStore(c)
	if !ok {
		return
	}

	now := time.Now()
	item := &models.TodoItem{
		ID:          primitive.NewObjectID().Hex(),
		Description: req.Description,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	project.Items[item.ID] = item
	project.OrderedIDs = append(project.OrderedIDs, item.ID)
	project.UpdatedAt = now

	if err := db.SaveTodoStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save todo item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *TodoHandler) UpdateTodoItem(c *gin.Context) {
	var req updateTodoItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	store, project, ok := h.loadTodoProjectStore(c)
	if !ok {
		return
	}

	item, exists := project.Items[c.Param("item_id")]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo item not found"})
		return
	}

	if req.Description != nil {
		item.Description = strings.TrimSpace(*req.Description)
	}
	if req.Status != nil {
		if !isValidTodoStatus(*req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		item.Status = *req.Status
	}
	if item.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	now := time.Now()
	item.UpdatedAt = now
	project.UpdatedAt = now

	if err := db.SaveTodoStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save todo item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *TodoHandler) DeleteTodoItem(c *gin.Context) {
	store, project, ok := h.loadTodoProjectStore(c)
	if !ok {
		return
	}

	id := c.Param("item_id")
	if _, exists := project.Items[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo item not found"})
		return
	}

	delete(project.Items, id)
	project.OrderedIDs = removeID(project.OrderedIDs, id)
	project.UpdatedAt = time.Now()

	if err := db.SaveTodoStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save todo item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *TodoHandler) loadTodoProject(c *gin.Context) (*models.TodoProject, bool) {
	_, project, ok := h.loadTodoProjectStore(c)
	return project, ok
}

func (h *TodoHandler) loadTodoProjectStore(c *gin.Context) (*models.TodoStore, *models.TodoProject, bool) {
	store, err := db.LoadTodoStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load todo projects"})
		return nil, nil, false
	}

	project, exists := store.Projects[c.Param("project_id")]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo project not found"})
		return nil, nil, false
	}

	return store, project, true
}

func isValidTodoStatus(status models.TodoStatus) bool {
	return status == models.TodoStatusPending || status == models.TodoStatusDone
}

func removeID(ids []string, target string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == target {
			continue
		}
		result = append(result, id)
	}
	return result
}
