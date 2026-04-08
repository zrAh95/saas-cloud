package handlers

import (
	"net/http"
	"os"
	"saas-cloud/config"
	"saas-cloud/models"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "File tidak ditemukan",
        })
        return
    }
	

    // simpan ke folder uploads
    filePath := "uploads/" + file.Filename
	c.SaveUploadedFile(file, filePath)

	// ambil info
	size := file.Size
	mimeType := file.Header.Get("Content-Type")

	// simpan ke DB
	newFile := models.File{
		UserID:     1, // sementara hardcode dulu
		FileName:   file.Filename,
		FilePath:   filePath,
		Size:       size,
		MimeType:   mimeType,
	}

	config.DB.Create(&newFile)

    err = c.SaveUploadedFile(file, filePath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Gagal menyimpan file",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Upload berhasil",
        "file_name": file.Filename,
        "path": filePath,
    })
}

func GetFiles(c *gin.Context) {
    files, err := os.ReadDir("uploads")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Gagal membaca folder",
        })
        return
    }

    var result []gin.H

    for _, file := range files {
        info, _ := file.Info()

        result = append(result, gin.H{
            "file_name": file.Name(),
            "size":      info.Size(),
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "files": result,
    })
}

func GetFileByID(c *gin.Context) {
    id := c.Param("id")

    filePath := "uploads/" + id

    // cek file ada
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "File tidak ditemukan",
        })
        return
    }

    // kirim file
    c.File(filePath)
}

func DeleteFile(c *gin.Context) {
    id := c.Param("id")

    filePath := "uploads/" + id

    err := os.Remove(filePath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Gagal menghapus file",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "File berhasil dihapus",
        "file": id,
    })
}
