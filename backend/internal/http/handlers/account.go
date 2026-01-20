package handler

import (
	"net/http"

	"banking-platform/internal/http/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccountHandler struct {
	accountService AccountService
}

func NewAccountHandler(accountService AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
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

	ctx := c.Request.Context()
	accounts, err := h.accountService.GetUserAccounts(ctx, userUUID)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	out := make([]*dto.AccountResponse, len(accounts))
	for i, a := range accounts {
		out[i] = &dto.AccountResponse{
			ID:           a.ID,
			Currency:     a.Currency,
			BalanceCents: a.BalanceCents,
		}
	}
	respondWithJSON(c, http.StatusOK, out)
}

func (h *AccountHandler) GetBalance(c *gin.Context) {
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

	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondWithError(c, "invalid account ID", http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	balance, err := h.accountService.GetAccountBalance(ctx, accountID, userUUID)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusOK, gin.H{"balance_cents": balance})
}

