package handlers

import (
	"errors"
	"saas-cloud/config"
	"saas-cloud/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getCurrentUser(c *gin.Context) (models.User, error) {
	userID := c.GetInt("user_id")

	var user models.User
	err := config.DB.First(&user, userID).Error
	return user, err
}

func hasAcceptedFileAccess(fileID uint, user models.User) (bool, error) {
	var access models.FileAccess
	err := config.DB.
		Where("file_id = ? AND target_email = ? AND status = ?", fileID, user.Email, "accepted").
		First(&access).Error
	if err != nil {
		return false, err
	}

	return true, nil
}

func canAccessItem(file models.File, user models.User) (bool, error) {
	if file.UserID == int(user.ID) {
		return true, nil
	}

	hasAccess, err := hasAcceptedFileAccess(file.ID, user)
	if err == nil {
		return hasAccess, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	currentParentID := file.ParentID
	for currentParentID != nil {
		var parent models.File
		if err := config.DB.First(&parent, *currentParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}

		if parent.UserID == int(user.ID) {
			return true, nil
		}

		hasParentAccess, err := hasAcceptedFileAccess(parent.ID, user)
		if err == nil && hasParentAccess {
			return true, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}

		currentParentID = parent.ParentID
	}

	return false, nil
}

func getAccessibleFolder(folderID uint, user models.User) (models.File, bool, error) {
	var folder models.File
	if err := config.DB.First(&folder, folderID).Error; err != nil {
		return folder, false, err
	}

	if !folder.IsFolder {
		return folder, false, nil
	}

	allowed, err := canAccessItem(folder, user)
	if err != nil {
		return folder, false, err
	}

	return folder, allowed, nil
}

func canModifyFolder(folder models.File, user models.User) bool {
	allowed, err := canAccessItem(folder, user)
	if err != nil {
		return false
	}
	return allowed
}
