package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func respondWithError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, gin.H{"error": message})
}

func respondWithJSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
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
