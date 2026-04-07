package handler

import "github.com/gin-gonic/gin"

// errorResponse writes a JSON error body and sets the HTTP status code.
func errorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"code": code, "message": message})
}
