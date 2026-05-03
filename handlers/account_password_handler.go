package handlers

import (
	"errors"
	"log"
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

const accountPasswordCollection = "account_passwords"

type AccountPasswordHandler struct{}

type createAccountPasswordRequest struct {
	Account     string `json:"account"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

type updateAccountPasswordRequest struct {
	Account     *string `json:"account"`
	Password    *string `json:"password"`
	Description *string `json:"description"`
}

func NewAccountPasswordHandler() *AccountPasswordHandler {
	return &AccountPasswordHandler{}
}

func (h *AccountPasswordHandler) ListAccountPasswords(c *gin.Context) {
	ctx := c.Request.Context()
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cur, err := db.DB().Collection(accountPasswordCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		log.Printf("[account_password] failed to list records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list account passwords"})
		return
	}
	defer cur.Close(ctx)

	result := make([]models.AccountPassword, 0)
	if err := cur.All(ctx, &result); err != nil {
		log.Printf("[account_password] failed to decode records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode account passwords"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AccountPasswordHandler) CreateAccountPassword(c *gin.Context) {
	var req createAccountPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[account_password] invalid create request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	record := models.AccountPassword{
		ID:          primitive.NewObjectID(),
		Account:     strings.TrimSpace(req.Account),
		Password:    strings.TrimSpace(req.Password),
		Description: strings.TrimSpace(req.Description),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if record.Account == "" || record.Password == "" || record.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account, password and description are required"})
		return
	}

	if _, err := db.DB().Collection(accountPasswordCollection).InsertOne(c.Request.Context(), record); err != nil {
		log.Printf("[account_password] failed to create record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account password"})
		return
	}

	log.Printf("[account_password] created record id=%s", record.ID.Hex())
	c.JSON(http.StatusOK, record)
}

func (h *AccountPasswordHandler) UpdateAccountPassword(c *gin.Context) {
	id, ok := parseAccountPasswordID(c)
	if !ok {
		return
	}

	var req updateAccountPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[account_password] invalid update request body id=%s: %v", id.Hex(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if req.Account != nil {
		set["account"] = strings.TrimSpace(*req.Account)
	}
	if req.Password != nil {
		set["password"] = strings.TrimSpace(*req.Password)
	}
	if req.Description != nil {
		set["description"] = strings.TrimSpace(*req.Description)
	}

	if account, exists := set["account"]; exists && account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account is required"})
		return
	}
	if password, exists := set["password"]; exists && password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}
	if description, exists := set["description"]; exists && description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	result, err := db.DB().Collection(accountPasswordCollection).UpdateByID(c.Request.Context(), id, bson.M{"$set": set})
	if err != nil {
		log.Printf("[account_password] failed to update record id=%s: %v", id.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account password"})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "account password not found"})
		return
	}

	record, ok := h.findAccountPassword(c, id)
	if !ok {
		return
	}

	log.Printf("[account_password] updated record id=%s", id.Hex())
	c.JSON(http.StatusOK, record)
}

func (h *AccountPasswordHandler) DeleteAccountPassword(c *gin.Context) {
	id, ok := parseAccountPasswordID(c)
	if !ok {
		return
	}

	result, err := db.DB().Collection(accountPasswordCollection).DeleteOne(c.Request.Context(), bson.M{"_id": id})
	if err != nil {
		log.Printf("[account_password] failed to delete record id=%s: %v", id.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account password"})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "account password not found"})
		return
	}

	log.Printf("[account_password] deleted record id=%s", id.Hex())
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func parseAccountPasswordID(c *gin.Context) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account password id"})
		return primitive.NilObjectID, false
	}
	return id, true
}

func (h *AccountPasswordHandler) findAccountPassword(c *gin.Context, id primitive.ObjectID) (models.AccountPassword, bool) {
	var record models.AccountPassword
	err := db.DB().Collection(accountPasswordCollection).
		FindOne(c.Request.Context(), bson.M{"_id": id}).
		Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"error": "account password not found"})
		return models.AccountPassword{}, false
	}
	if err != nil {
		log.Printf("[account_password] failed to load record id=%s: %v", id.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load account password"})
		return models.AccountPassword{}, false
	}

	return record, true
}
