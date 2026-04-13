package handlers

import (
	"net/http"
	"os"
	"saas-cloud/config"
	"saas-cloud/models"
	"saas-cloud/utils"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
		
	}
	
	userID := userIDInterface.(int)

	// 1. ambil file dulu
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "File tidak ditemukan")
		return
	}

	// 2. validasi tipe file
	allowedTypes := []string{"image/jpeg", "image/png", "application/pdf"}
	fileType := file.Header.Get("Content-Type")

	isValid := false
	for _, t := range allowedTypes {
		if t == fileType {
			isValid = true
			break
		}
	}

	if !isValid {
		utils.Error(c, http.StatusBadRequest, "Tipe file tidak diizinkan")
		return
	}

	// 3. limit size
	if file.Size > 5*1024*1024 {
		utils.Error(c, http.StatusBadRequest, "File terlalu besar")
		return
	}

	// 4. rename file (AMAN)
	fileName := uuid.New().String() + "_" + file.Filename
	filePath := "uploads/" + fileName

	// 5. simpan file
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal upload file")
		return
	}

	// 6. simpan ke DB
	newFile := models.File{
		UserID:       userID,
		FileName:     fileName, // <- ganti ini juga biar konsisten
		OriginalName: file.Filename,
		FilePath:     filePath,
		Size:         file.Size,
		MimeType:     fileType,
	}

	config.DB.Create(&newFile)

	utils.Success(c, http.StatusOK, "Upload berhasil", newFile)
}

func GetFiles(c *gin.Context) {

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	userID := userIDInterface.(int)

	var files []models.File

	result := config.DB.Where("user_id = ?", userID).Find(&files)

	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, result.Error.Error())
		return
	}

	utils.Success(c, http.StatusOK, "List file", files)
}

func GetFileByID(c *gin.Context) {
	userID := c.GetInt("user_id")
	id := c.Param("id")

	var file models.File

	if err := config.DB.First(&file, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
		return
	}

	// SECURITY
	if file.UserID != userID {
		utils.Error(c, http.StatusForbidden, "Akses ditolak")
		return
	}

	c.File(file.FilePath)
}

func DeleteFile(c *gin.Context) {
    userID := c.GetInt("user_id")
    id := c.Param("id")

    var file models.File

    // ambil file
    if err := config.DB.First(&file, id).Error; err != nil {
        utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
        return
    }

    // 🔒 SECURITY: cek ownership
    if file.UserID != userID {
        utils.Error(c, http.StatusForbidden, "Akses ditolak")
        return
    }

    // hapus file fisik
    if err := os.Remove(file.FilePath); err != nil {
        utils.Error(c, http.StatusInternalServerError, "Gagal hapus file fisik")
        return
    }

    // hapus dari DB
    config.DB.Delete(&file)

    utils.Success(c, http.StatusOK, "File berhasil dihapus", nil)
}