package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"banking-platform/internal/apperr"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func respondWithError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, gin.H{"error": message})
}

func respondWithJSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
}

func respondWithServiceError(c *gin.Context, err error) {
	var pub *apperr.PublicError
	if errors.As(err, &pub) && pub != nil {
		slog.Default().Warn(
			"request failed (public error)",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"error", err,
			"cause", apperr.RootCause(err),
		)
		respondWithError(c, pub.Message, pub.Status)
		return
	}

	cause := apperr.RootCause(err)

	isClientError :=
		errors.Is(cause, apperr.ErrUserExists) ||
			errors.Is(cause, apperr.ErrInvalidCredentials) ||
			errors.Is(cause, apperr.ErrInvalidToken) ||
			errors.Is(cause, apperr.ErrUnauthorized) ||
			errors.Is(cause, apperr.ErrAccountNotFound) ||
			errors.Is(cause, apperr.ErrTransactionNotFound) ||
			errors.Is(cause, apperr.ErrUserNotFound) ||
			errors.Is(cause, apperr.ErrInsufficientFunds) ||
			errors.Is(cause, apperr.ErrInvalidAmount) ||
			errors.Is(cause, apperr.ErrInvalidCurrency) ||
			errors.Is(cause, apperr.ErrCurrenciesMustDiffer) ||
			errors.Is(cause, apperr.ErrCannotTransferToSelf) ||
			errors.Is(cause, apperr.ErrLiquidityUnavailable)

	if isClientError {
		slog.Default().Warn(
			"request failed",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"error", err,
			"cause", cause,
		)
	} else {
		slog.Default().Error(
			"request failed",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"error", err,
			"cause", cause,
		)
	}

	switch {
	case errors.Is(cause, apperr.ErrUserExists):
		respondWithError(c, apperr.ErrUserExists.Error(), http.StatusConflict)
	case errors.Is(cause, apperr.ErrInvalidCredentials):
		respondWithError(c, apperr.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
	case errors.Is(cause, apperr.ErrInvalidToken):
		respondWithError(c, apperr.ErrInvalidToken.Error(), http.StatusUnauthorized)
	case errors.Is(cause, apperr.ErrUnauthorized):
		respondWithError(c, apperr.ErrUnauthorized.Error(), http.StatusForbidden)
	case errors.Is(cause, apperr.ErrAccountNotFound):
		respondWithError(c, apperr.ErrAccountNotFound.Error(), http.StatusNotFound)
	case errors.Is(cause, apperr.ErrTransactionNotFound):
		respondWithError(c, apperr.ErrTransactionNotFound.Error(), http.StatusNotFound)
	case errors.Is(cause, apperr.ErrUserNotFound):
		if c.FullPath() == "/transactions/transfer" {
			respondWithError(c, "recipient not found", http.StatusBadRequest)
			return
		}
		respondWithError(c, apperr.ErrUserNotFound.Error(), http.StatusNotFound)
	case errors.Is(cause, apperr.ErrInsufficientFunds):
		respondWithError(c, apperr.ErrInsufficientFunds.Error(), http.StatusBadRequest)
	case errors.Is(cause, apperr.ErrInvalidAmount):
		respondWithError(c, apperr.ErrInvalidAmount.Error(), http.StatusBadRequest)
	case errors.Is(cause, apperr.ErrInvalidCurrency):
		respondWithError(c, apperr.ErrInvalidCurrency.Error(), http.StatusBadRequest)
	case errors.Is(cause, apperr.ErrCurrenciesMustDiffer):
		respondWithError(c, apperr.ErrCurrenciesMustDiffer.Error(), http.StatusBadRequest)
	case errors.Is(cause, apperr.ErrCannotTransferToSelf):
		respondWithError(c, apperr.ErrCannotTransferToSelf.Error(), http.StatusBadRequest)
	case errors.Is(cause, apperr.ErrLiquidityUnavailable):
		respondWithError(c, apperr.ErrLiquidityUnavailable.Error(), http.StatusConflict)
	default:
		respondWithError(c, "internal_error", http.StatusInternalServerError)
	}
}

type validationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func respondWithBindError(c *gin.Context, err error) {
	if err == nil {
		respondWithError(c, "invalid request", http.StatusBadRequest)
		return
	}

	if errors.Is(err, io.EOF) {
		respondWithError(c, "request body is required", http.StatusBadRequest)
		return
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		respondWithError(c, "invalid json", http.StatusBadRequest)
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "invalid character") || strings.Contains(msg, "unexpected eof") || strings.Contains(msg, "unexpected end of json") {
		respondWithError(c, "invalid json", http.StatusBadRequest)
		return
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		respondWithError(c, "invalid field type: "+toSnakeCase(typeErr.Field), http.StatusBadRequest)
		return
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]validationFieldError, 0, len(ve))
		for _, fe := range ve {
			field := toSnakeCase(fe.Field())
			out = append(out, validationFieldError{
				Field:   field,
				Message: validationMessage(fe),
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": out})
		return
	}

	respondWithError(c, "invalid request body", http.StatusBadRequest)
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	case "gt":
		return "must be greater than " + fe.Param()
	default:
		return "is invalid"
	}
}

func toSnakeCase(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteByte(byte(r + ('a' - 'A')))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

