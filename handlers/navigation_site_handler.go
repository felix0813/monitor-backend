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

type NavigationSiteHandler struct{}

type createNavigationSiteRequest struct {
	ImageURL    string   `json:"image_url"`
	URL         string   `json:"url"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type updateNavigationSiteRequest struct {
	ImageURL    *string   `json:"image_url"`
	URL         *string   `json:"url"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Tags        *[]string `json:"tags"`
}

type reorderNavigationSitesRequest struct {
	IDs []string `json:"ids"`
}

func NewNavigationSiteHandler() *NavigationSiteHandler {
	return &NavigationSiteHandler{}
}

func (h *NavigationSiteHandler) ListNavigationSites(c *gin.Context) {
	store, err := db.LoadNavigationSiteStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation sites"})
		return
	}

	result := make([]*models.NavigationSite, 0, len(store.OrderedIDs))
	for _, id := range store.OrderedIDs {
		site, exists := store.Sites[id]
		if !exists {
			continue
		}
		result = append(result, site)
	}

	c.JSON(http.StatusOK, result)
}

func (h *NavigationSiteHandler) CreateNavigationSite(c *gin.Context) {
	var req createNavigationSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	req.Name = strings.TrimSpace(req.Name)
	req.ImageURL = strings.TrimSpace(req.ImageURL)
	req.Description = strings.TrimSpace(req.Description)

	if req.URL == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url and name are required"})
		return
	}

	store, err := db.LoadNavigationSiteStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation sites"})
		return
	}

	now := time.Now()
	site := &models.NavigationSite{
		ID:          primitive.NewObjectID().Hex(),
		ImageURL:    req.ImageURL,
		URL:         req.URL,
		Name:        req.Name,
		Description: req.Description,
		Tags:        normalizeTags(req.Tags),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	store.Sites[site.ID] = site
	store.OrderedIDs = append(store.OrderedIDs, site.ID)

	if err := db.SaveNavigationSiteStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save navigation site"})
		return
	}

	c.JSON(http.StatusOK, site)
}

func (h *NavigationSiteHandler) UpdateNavigationSite(c *gin.Context) {
	var req updateNavigationSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	store, err := db.LoadNavigationSiteStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation sites"})
		return
	}

	site, exists := store.Sites[c.Param("id")]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "navigation site not found"})
		return
	}

	if req.ImageURL != nil {
		site.ImageURL = strings.TrimSpace(*req.ImageURL)
	}
	if req.URL != nil {
		site.URL = strings.TrimSpace(*req.URL)
	}
	if req.Name != nil {
		site.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		site.Description = strings.TrimSpace(*req.Description)
	}
	if req.Tags != nil {
		site.Tags = normalizeTags(*req.Tags)
	}

	if site.URL == "" || site.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url and name are required"})
		return
	}

	site.UpdatedAt = time.Now()

	if err := db.SaveNavigationSiteStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save navigation site"})
		return
	}

	c.JSON(http.StatusOK, site)
}

func (h *NavigationSiteHandler) DeleteNavigationSite(c *gin.Context) {
	store, err := db.LoadNavigationSiteStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation sites"})
		return
	}

	id := c.Param("id")
	if _, exists := store.Sites[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "navigation site not found"})
		return
	}

	delete(store.Sites, id)
	store.OrderedIDs = removeNavigationSiteID(store.OrderedIDs, id)

	if err := db.SaveNavigationSiteStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save navigation site"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *NavigationSiteHandler) ReorderNavigationSites(c *gin.Context) {
	var req reorderNavigationSitesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	store, err := db.LoadNavigationSiteStore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load navigation sites"})
		return
	}

	if len(req.IDs) != len(store.Sites) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids count does not match navigation sites count"})
		return
	}

	seen := make(map[string]struct{}, len(req.IDs))
	orderedIDs := make([]string, 0, len(req.IDs))
	for _, id := range req.IDs {
		id = strings.TrimSpace(id)
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id cannot be empty"})
			return
		}
		if _, exists := store.Sites[id]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contains unknown navigation site id"})
			return
		}
		if _, exists := seen[id]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate navigation site id"})
			return
		}
		seen[id] = struct{}{}
		orderedIDs = append(orderedIDs, id)
	}

	store.OrderedIDs = orderedIDs
	if err := db.SaveNavigationSiteStore(store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save navigation site order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		result = append(result, tag)
	}

	return result
}

func removeNavigationSiteID(ids []string, target string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == target {
			continue
		}
		result = append(result, id)
	}
	return result
}
