package handler

import (
	"errors"
	"net/http"
	"strconv"

	"banking-platform/internal/apperr"
	"banking-platform/internal/model"
	"banking-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	transactionService service.ITransactionService
}

func NewTransactionHandler(transactionService service.ITransactionService) *TransactionHandler {
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

	var req model.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	transaction, err := h.transactionService.Transfer(ctx, userUUID, &req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, apperr.ErrInsufficientFunds) {
			statusCode = http.StatusBadRequest
		}
		respondWithError(c, err.Error(), statusCode)
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

	var req model.ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	transaction, err := h.transactionService.Exchange(ctx, userUUID, &req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, apperr.ErrInsufficientFunds) {
			statusCode = http.StatusBadRequest
		}
		respondWithError(c, err.Error(), statusCode)
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

	filter := &model.TransactionFilter{
		Type: model.TransactionType(c.Query("type")),
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
		respondWithError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(c, http.StatusOK, transactions)
}
