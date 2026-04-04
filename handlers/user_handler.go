package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int)

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile",
		"user_id": userID,
	})
}