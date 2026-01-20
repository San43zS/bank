package handler

import (
	"net/http"

	"banking-platform/internal/domain"
	"banking-platform/internal/http/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	ctx := c.Request.Context()
	out, err := h.authService.Register(ctx, &domain.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusCreated, &dto.AuthResponse{
		AccessToken:  out.Tokens.AccessToken,
		RefreshToken: out.Tokens.RefreshToken,
		User: &dto.UserResponse{
			ID:        out.User.ID,
			Email:     out.User.Email,
			FirstName: out.User.FirstName,
			LastName:  out.User.LastName,
			CreatedAt: out.User.CreatedAt,
			UpdatedAt: out.User.UpdatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	ctx := c.Request.Context()
	out, err := h.authService.Login(ctx, &domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusOK, &dto.AuthResponse{
		AccessToken:  out.Tokens.AccessToken,
		RefreshToken: out.Tokens.RefreshToken,
		User: &dto.UserResponse{
			ID:        out.User.ID,
			Email:     out.User.Email,
			FirstName: out.User.FirstName,
			LastName:  out.User.LastName,
			CreatedAt: out.User.CreatedAt,
			UpdatedAt: out.User.UpdatedAt,
		},
	})
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
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusOK, &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	ctx := c.Request.Context()
	tokenPair, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusOK, &dto.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.Logout(ctx, &domain.LogoutInput{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
	}); err != nil {
		respondWithServiceError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

