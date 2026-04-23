package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"saas-cloud/config"
	"saas-cloud/models"
	"saas-cloud/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateFolderRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parent_id"`
}

func UploadFile(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "File tidak ditemukan")
		return
	}

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

	if file.Size > 5*1024*1024 {
		utils.Error(c, http.StatusBadRequest, "File terlalu besar")
		return
	}

	parentID, err := parseOptionalParentID(c.PostForm("parent_id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "parent_id tidak valid")
		return
	}

	if parentID != nil {
		parentFolder, allowed, err := getAccessibleFolder(*parentID, user)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Error(c, http.StatusNotFound, "Folder parent tidak ditemukan")
				return
			}
			utils.Error(c, http.StatusInternalServerError, "Gagal cek folder parent")
			return
		}
		if !allowed {
			utils.Error(c, http.StatusForbidden, "Tidak punya akses ke folder parent")
			return
		}
		if !canModifyFolder(parentFolder, user) {
			utils.Error(c, http.StatusForbidden, "Folder share hanya bisa dibaca")
			return
		}
	}

	fileName := uuid.New().String() + "_" + filepath.Base(file.Filename)
	filePath := "uploads/" + fileName

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal upload file")
		return
	}

	newFile := models.File{
		UserID:       int(user.ID),
		ParentID:     parentID,
		FileName:     fileName,
		OriginalName: file.Filename,
		FilePath:     filePath,
		Size:         file.Size,
		MimeType:     fileType,
		IsFolder:     false,
	}

	if err := config.DB.Create(&newFile).Error; err != nil {
		_ = os.Remove(filePath)
		utils.Error(c, http.StatusInternalServerError, "Gagal simpan data file")
		return
	}

	utils.Success(c, http.StatusOK, "Upload berhasil", newFile)
}

func CreateFolder(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Input tidak valid")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		utils.Error(c, http.StatusBadRequest, "Nama folder wajib diisi")
		return
	}

	if strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") {
		utils.Error(c, http.StatusBadRequest, "Nama folder tidak valid")
		return
	}

	if req.ParentID != nil {
		parentFolder, allowed, err := getAccessibleFolder(*req.ParentID, user)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Error(c, http.StatusNotFound, "Folder parent tidak ditemukan")
				return
			}
			utils.Error(c, http.StatusInternalServerError, "Gagal cek folder parent")
			return
		}
		if !allowed {
			utils.Error(c, http.StatusForbidden, "Tidak punya akses ke folder parent")
			return
		}
		if !canModifyFolder(parentFolder, user) {
			utils.Error(c, http.StatusForbidden, "Folder share hanya bisa dibaca")
			return
		}
	}

	var existing models.File
	query := config.DB.Where("original_name = ? AND is_folder = ?", req.Name, true)
	if req.ParentID == nil {
		query = query.Where("user_id = ? AND parent_id IS NULL", user.ID)
	} else {
		query = query.Where("parent_id = ?", *req.ParentID)
	}

	err = query.First(&existing).Error
	if err == nil {
		utils.Error(c, http.StatusBadRequest, "Folder dengan nama itu sudah ada")
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal cek folder")
		return
	}

	folder := models.File{
		UserID:       int(user.ID),
		ParentID:     req.ParentID,
		FileName:     req.Name,
		OriginalName: req.Name,
		FilePath:     "",
		Size:         0,
		MimeType:     "folder",
		IsFolder:     true,
	}

	if err := config.DB.Create(&folder).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal membuat folder")
		return
	}

	utils.Success(c, http.StatusOK, "Folder berhasil dibuat", folder)
}

func GetFiles(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	parentID, err := parseOptionalParentID(c.Query("parent_id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "parent_id tidak valid")
		return
	}

	if parentID != nil {
		folder, allowed, err := getAccessibleFolder(*parentID, user)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Error(c, http.StatusNotFound, "Folder tidak ditemukan")
				return
			}
			utils.Error(c, http.StatusInternalServerError, "Gagal cek folder")
			return
		}
		if !allowed {
			utils.Error(c, http.StatusForbidden, "Tidak punya akses ke folder ini")
			return
		}

		var files []models.File
		query := config.DB.Where("parent_id = ?", folder.ID)
		if folder.UserID != int(user.ID) {
			query = query.Where("user_id = ?", folder.UserID)
		}
		if err := query.Order("is_folder DESC, created_at DESC").Find(&files).Error; err != nil {
			utils.Error(c, http.StatusInternalServerError, "Gagal mengambil isi folder")
			return
		}

		utils.Success(c, http.StatusOK, "List file", files)
		return
	}

	var files []models.File
	result := config.DB.
		Table("tb_files f").
		Select("DISTINCT f.id, f.user_id, f.parent_id, f.filename, f.original_name, f.path, f.size, f.mime_type, f.is_folder, f.created_at").
		Joins("LEFT JOIN tb_file_access fa ON fa.file_id = f.id").
		Where("f.parent_id IS NULL").
		Where("f.user_id = ? OR (fa.target_email = ? AND fa.status = ?)", user.ID, user.Email, "accepted").
		Order("f.is_folder DESC, f.created_at DESC").
		Scan(&files)
	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, result.Error.Error())
		return
	}

	utils.Success(c, http.StatusOK, "List file", files)
}

func GetFolderContents(c *gin.Context) {
	c.Request.URL.RawQuery = "parent_id=" + c.Param("id")
	GetFiles(c)
}

func GetFileByID(c *gin.Context) {
	id := c.Param("id")

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
		return
	}

	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	allowed, err := canAccessItem(file, user)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal cek akses file")
		return
	}
	if !allowed {
		utils.Error(c, http.StatusForbidden, "Akses ditolak")
		return
	}

	if file.IsFolder {
		utils.Success(c, http.StatusOK, "Detail folder", file)
		return
	}

	if _, err := os.Stat(file.FilePath); err != nil {
		utils.Error(c, http.StatusNotFound, "File fisik tidak ditemukan")
		return
	}

	c.File(file.FilePath)
}

func DeleteFile(c *gin.Context) {
	userID := c.GetInt("user_id")
	id := c.Param("id")

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
		return
	}

	if file.UserID != userID {
		utils.Error(c, http.StatusForbidden, "Akses ditolak")
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal memulai transaksi")
		return
	}

	if err := deleteFileNode(tx, file); err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal menyimpan penghapusan file")
		return
	}

	if file.IsFolder {
		utils.Success(c, http.StatusOK, "Folder berhasil dihapus", nil)
		return
	}

	utils.Success(c, http.StatusOK, "File berhasil dihapus", nil)
}

func DownloadFile(c *gin.Context) {
	fileID := c.Param("id")

	var file models.File
	if err := config.DB.First(&file, fileID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "File tidak ditemukan")
		return
	}

	if file.IsFolder {
		utils.Error(c, http.StatusBadRequest, "Folder tidak bisa didownload langsung")
		return
	}

	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	allowed, err := canAccessItem(file, user)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal cek akses file")
		return
	}
	if !allowed {
		utils.Error(c, http.StatusForbidden, "Tidak punya akses ke file ini")
		return
	}

	if _, err := os.Stat(file.FilePath); err != nil {
		utils.Error(c, http.StatusNotFound, "File fisik tidak ditemukan")
		return
	}

	c.FileAttachment(file.FilePath, file.OriginalName)
}

func parseOptionalParentID(raw string) (*uint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil, err
	}

	parentID := uint(id)
	return &parentID, nil
}

func deleteFileNode(tx *gorm.DB, file models.File) error {
	if file.IsFolder {
		var children []models.File
		if err := tx.Where("parent_id = ?", file.ID).Find(&children).Error; err != nil {
			return errors.New("gagal mengambil isi folder")
		}

		for _, child := range children {
			if err := deleteFileNode(tx, child); err != nil {
				return err
			}
		}
	} else if file.FilePath != "" {
		if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
			return errors.New("gagal hapus file fisik")
		}
	}

	if err := tx.Where("file_id = ?", file.ID).Delete(&models.FileAccess{}).Error; err != nil {
		return errors.New("gagal hapus data akses file")
	}

	if err := tx.Delete(&file).Error; err != nil {
		return errors.New("gagal hapus data file")
	}

	return nil
}
