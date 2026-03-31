package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShareFile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Share file",
	})
}

func GetSharedFiles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get shared files",
	})
}

func AcceptShare(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Accept share",
	})
}
