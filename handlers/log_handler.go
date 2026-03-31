package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get logs",
	})
}
