package handler

import (
	"github.com/gin-gonic/gin"
)

func respondWithError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, gin.H{"error": message})
}

func respondWithJSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
}
