package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload file",
	})
}

func GetFiles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get all files",
	})
}

func GetFileByID(c *gin.Context) {
	id := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"message": "Get file by ID",
		"id":      id,
	})
}

func DeleteFile(c *gin.Context) {
	id := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"message": "Delete file",
		"id":      id,
	})
}
