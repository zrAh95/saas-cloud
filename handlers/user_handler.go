package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"saas-cloud/config"
	"saas-cloud/models"
	"saas-cloud/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UpdateProfileInput struct {
	Name            *string `json:"name"`
	Email           *string `json:"email"`
	Phone           *string `json:"phone"`
	CurrentPassword string  `json:"current_password"`
	NewPassword     string  `json:"new_password"`
}

func GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	utils.Success(c, http.StatusOK, "User profile", gin.H{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phone":       user.Phone,
		"has_avatar":  user.AvatarPath != "",
		"is_verified": user.IsVerified,
		"created_at":  user.CreatedAt,
	})
}

func UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int)

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Input tidak valid")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	updates := map[string]interface{}{}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if len(name) > 100 {
			utils.Error(c, http.StatusBadRequest, "Nama maksimal 100 karakter")
			return
		}
		updates["name"] = name
	}

	emailChanged := false
	if input.Email != nil {
		email := strings.TrimSpace(*input.Email)
		if email == "" {
			utils.Error(c, http.StatusBadRequest, "Email wajib diisi")
			return
		}
		if !utils.ValidateEmail(email) {
			utils.Error(c, http.StatusBadRequest, "Format email tidak valid")
			return
		}
		if email != user.Email {
			emailChanged = true
			var existing models.User
			err := config.DB.Where("email = ? AND id <> ?", email, user.ID).First(&existing).Error
			if err == nil {
				utils.Error(c, http.StatusBadRequest, "Email sudah terdaftar")
				return
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Error(c, http.StatusInternalServerError, "Gagal cek email")
				return
			}
			updates["email"] = email
		}
	}

	phoneChanged := false
	if input.Phone != nil {
		phone := utils.NormalizeWhatsAppNumber(*input.Phone)
		if phone != "" && !utils.ValidateWhatsAppNumber(phone) {
			utils.Error(c, http.StatusBadRequest, "Format nomor WhatsApp tidak valid")
			return
		}
		if phone != user.Phone {
			phoneChanged = true
			updates["phone"] = phone
		}
	}

	newPassword := strings.TrimSpace(input.NewPassword)
	passwordChanged := newPassword != ""
	if emailChanged || phoneChanged || passwordChanged {
		if input.CurrentPassword == "" {
			utils.Error(c, http.StatusBadRequest, "Password lama wajib untuk ganti email/nomor/password")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
			utils.Error(c, http.StatusUnauthorized, "Password lama salah")
			return
		}
	}

	if passwordChanged {
		passwordErrors := utils.ValidatePassword(newPassword)
		for _, msg := range []string{
			passwordErrors["length"],
			passwordErrors["uppercase"],
			passwordErrors["lowercase"],
			passwordErrors["number"],
		} {
			if msg != "" {
				utils.Error(c, http.StatusBadRequest, "Password tidak memenuhi kriteria: "+msg)
				return
			}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Gagal hash password baru")
			return
		}
		updates["password"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		utils.Success(c, http.StatusOK, "Tidak ada perubahan profile", gin.H{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"phone":       user.Phone,
			"has_avatar":  user.AvatarPath != "",
			"is_verified": user.IsVerified,
			"created_at":  user.CreatedAt,
		})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal update profile")
		return
	}

	if err := config.DB.First(&user, user.ID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil profile terbaru")
		return
	}

	utils.Success(c, http.StatusOK, "Profile berhasil diupdate", gin.H{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phone":       user.Phone,
		"has_avatar":  user.AvatarPath != "",
		"is_verified": user.IsVerified,
		"created_at":  user.CreatedAt,
	})
}

func GetProfileAvatar(c *gin.Context) {
	userID := c.MustGet("user_id").(int)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	if user.AvatarPath == "" {
		utils.Error(c, http.StatusNotFound, "Foto profile belum ada")
		return
	}

	if _, err := os.Stat(user.AvatarPath); err != nil {
		utils.Error(c, http.StatusNotFound, "File foto profile tidak ditemukan")
		return
	}

	c.File(user.AvatarPath)
}

func UploadProfileAvatar(c *gin.Context) {
	userID := c.MustGet("user_id").(int)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Foto profile tidak ditemukan")
		return
	}

	if file.Size > 2*1024*1024 {
		utils.Error(c, http.StatusBadRequest, "Foto profile maksimal 2MB")
		return
	}

	fileType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedTypes[fileType] {
		utils.Error(c, http.StatusBadRequest, "Foto profile harus JPG, PNG, atau WEBP")
		return
	}

	if err := os.MkdirAll("uploads/avatars", 0755); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal menyiapkan folder avatar")
		return
	}

	fileName := uuid.New().String() + "_" + filepath.Base(file.Filename)
	filePath := filepath.Join("uploads", "avatars", fileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal upload foto profile")
		return
	}

	oldAvatarPath := user.AvatarPath
	if err := config.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("avatar_path", filePath).Error; err != nil {
		_ = os.Remove(filePath)
		utils.Error(c, http.StatusInternalServerError, "Gagal menyimpan foto profile")
		return
	}

	if oldAvatarPath != "" && oldAvatarPath != filePath {
		_ = os.Remove(oldAvatarPath)
	}

	utils.Success(c, http.StatusOK, "Foto profile berhasil diupdate", gin.H{
		"has_avatar": true,
	})
}
