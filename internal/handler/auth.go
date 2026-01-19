package handler

import (
	"errors"
	"net/http"

	"banking-platform/internal/apperr"
	"banking-platform/internal/model"
	"banking-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.IAuthService
}

func NewAuthHandler(authService service.IAuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	response, err := h.authService.Register(ctx, &req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, apperr.ErrUserExists) {
			statusCode = http.StatusConflict
		}
		respondWithError(c, err.Error(), statusCode)
		return
	}

	respondWithJSON(c, http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	response, err := h.authService.Login(ctx, &req)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if errors.Is(err, apperr.ErrInvalidCredentials) {
			statusCode = http.StatusUnauthorized
		}
		respondWithError(c, err.Error(), statusCode)
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
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
	user, err := h.authService.GetUserByID(ctx, userUUID)
	if err != nil {
		statusCode := http.StatusNotFound
		if errors.Is(err, apperr.ErrUserNotFound) {
			statusCode = http.StatusNotFound
		}
		respondWithError(c, err.Error(), statusCode)
		return
	}

	respondWithJSON(c, http.StatusOK, user)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	tokenPair, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if errors.Is(err, apperr.ErrInvalidToken) {
			statusCode = http.StatusUnauthorized
		}
		respondWithError(c, err.Error(), statusCode)
		return
	}

	respondWithJSON(c, http.StatusOK, tokenPair)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req model.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.Logout(ctx, req.AccessToken, req.RefreshToken); err != nil {
		respondWithError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
