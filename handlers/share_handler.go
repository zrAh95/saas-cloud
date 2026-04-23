package handlers

import (
	"errors"
	"net/http"
	"saas-cloud/config"
	"saas-cloud/models"
	"saas-cloud/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ShareRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RevokeAccessRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type SharedFileItem struct {
	ShareID      uint   `json:"share_id"`
	FileID       int    `json:"file_id"`
	OwnerID      int    `json:"owner_id"`
	OwnerEmail   string `json:"owner_email"`
	TargetEmail  string `json:"target_email"`
	Status       string `json:"status"`
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FilePath     string `json:"file_path"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size"`
	IsFolder     bool   `json:"is_folder"`
}

func ShareFile(c *gin.Context) {
	userID := c.GetInt("user_id")
	fileIDParam := c.Param("id")

	fileID, err := strconv.Atoi(fileIDParam)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var req ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Input tidak valid")
		return
	}

	var file models.File
	if err := config.DB.First(&file, fileID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
		return
	}

	if file.UserID != userID {
		utils.Error(c, http.StatusForbidden, "Akses ditolak")
		return
	}

	var targetUser models.User
	if err := config.DB.Where("email = ?", req.Email).First(&targetUser).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	if userID == int(targetUser.ID) {
		utils.Error(c, http.StatusBadRequest, "Tidak bisa share ke diri sendiri")
		return
	}

	var existing models.FileAccess
	err = config.DB.Where("file_id = ? AND target_email = ?", file.ID, req.Email).First(&existing).Error
	if err == nil {
		utils.Error(c, http.StatusBadRequest, "File sudah di-share ke user ini")
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal cek data share")
		return
	}

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
	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	var sharedFiles []SharedFileItem
	err = config.DB.
		Table("tb_file_access fa").
		Select(`
			fa.id AS share_id,
			fa.file_id,
			fa.owner_id,
			owner.email AS owner_email,
			fa.target_email,
			fa.status,
			f.filename AS file_name,
			f.original_name,
			f.path AS file_path,
			f.mime_type,
			f.size,
			f.is_folder
		`).
		Joins("JOIN tb_files f ON f.id = fa.file_id").
		Joins("JOIN tb_users owner ON owner.id = fa.owner_id").
		Where("fa.target_email = ?", user.Email).
		Order("fa.id DESC").
		Scan(&sharedFiles).Error
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil file share")
		return
	}

	utils.Success(c, http.StatusOK, "List file share", sharedFiles)
}

func AcceptShare(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	shareID := c.Param("id")

	var access models.FileAccess
	if err := config.DB.First(&access, shareID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Data share tidak ditemukan")
		return
	}

	if access.TargetEmail != user.Email {
		utils.Error(c, http.StatusForbidden, "Kamu tidak punya akses ke share ini")
		return
	}

	if access.Status != "pending" {
		utils.Error(c, http.StatusBadRequest, "Share sudah diproses")
		return
	}

	access.Status = "accepted"
	if err := config.DB.Save(&access).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal menerima share")
		return
	}

	utils.Success(c, http.StatusOK, "Berhasil menerima file", access)
}

func RevokeFileAccess(c *gin.Context) {
	userID := c.GetInt("user_id")
	fileID := c.Param("id")

	var req RevokeAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Input tidak valid")
		return
	}

	var file models.File
	if err := config.DB.First(&file, fileID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
		return
	}

	if file.UserID != userID {
		utils.Error(c, http.StatusForbidden, "Akses ditolak")
		return
	}

	result := config.DB.Where("file_id = ? AND target_email = ?", file.ID, req.Email).Delete(&models.FileAccess{})
	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal revoke akses file")
		return
	}

	if result.RowsAffected == 0 {
		utils.Error(c, http.StatusNotFound, "Akses user untuk file ini tidak ditemukan")
		return
	}

	utils.Success(c, http.StatusOK, "Akses file berhasil dicabut", gin.H{
		"file_id": file.ID,
		"email":   req.Email,
	})
}
