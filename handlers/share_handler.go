package handlers

import (
	"net/http"
	"saas-cloud/config"
	"saas-cloud/models"
	"saas-cloud/utils"

	"github.com/gin-gonic/gin"
)

type ShareRequest struct {
    Email string `json:"email" binding:"required,email"`
}

func ShareFile(c *gin.Context) {
    userID := c.GetInt("user_id")
    fileID := c.Param("id")

    var req ShareRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error(c, http.StatusBadRequest, "Input tidak valid")
        return
    }

    // ambil file
    var file models.File
    if err := config.DB.First(&file, fileID).Error; err != nil {
        utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
        return
    }

    // cek ownership
    if file.UserID != userID {
        utils.Error(c, http.StatusForbidden, "Akses ditolak")
        return
    }

    // cari user target
    var targetUser models.User
    if err := config.DB.Where("email = ?", req.Email).First(&targetUser).Error; err != nil {
        utils.Error(c, http.StatusNotFound, "User tidak ditemukan")
        return
    }

    // insert ke tb_file_access
    access := models.FileAccess{
        FileID:      int(file.ID),
        OwnerID:     userID,
        TargetEmail: req.Email,
        Status:      "pending",
    }

    if err := config.DB.Create(&access).Error; err != nil {
        utils.Error(c, http.StatusInternalServerError, "Gagal share file")
        return
    }

    utils.Success(c, http.StatusOK, "File berhasil di-share", access)
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
