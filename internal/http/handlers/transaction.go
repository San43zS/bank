package handler

import (
	"net/http"
	"strconv"
	"strings"

	"banking-platform/internal/domain"
	"banking-platform/internal/http/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	transactionService TransactionService
}

func NewTransactionHandler(transactionService TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

func (h *TransactionHandler) Transfer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondWithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		respondWithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}
	if (req.ToUserID == nil) == (req.ToUserEmail == nil) {
		respondWithJSON(c, http.StatusBadRequest, gin.H{"error": "validation_error", "fields": []validationFieldError{{Field: "to_user_id", Message: "provide either to_user_id or to_user_email"}, {Field: "to_user_email", Message: "provide either to_user_id or to_user_email"}}})
		return
	}
	if req.ToUserEmail != nil && strings.TrimSpace(*req.ToUserEmail) == "" {
		respondWithJSON(c, http.StatusBadRequest, gin.H{"error": "validation_error", "fields": []validationFieldError{{Field: "to_user_email", Message: "cannot be empty"}}})
		return
	}

	ctx := c.Request.Context()
	transaction, err := h.transactionService.Transfer(ctx, userUUID, &req)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusCreated, transaction)
}

func (h *TransactionHandler) Exchange(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondWithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		respondWithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	var req dto.ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	ctx := c.Request.Context()
	transaction, err := h.transactionService.Exchange(ctx, userUUID, &req)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusCreated, transaction)
}

func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondWithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		respondWithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	filter := &dto.TransactionFilter{
		Type: domain.TransactionType(c.Query("type")),
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}

	filter.Page = page
	filter.Limit = limit

	ctx := c.Request.Context()
	transactions, err := h.transactionService.GetUserTransactions(ctx, userUUID, filter)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusOK, transactions)
}

