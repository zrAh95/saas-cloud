package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"saas-cloud/config"
	"saas-cloud/models"
	"saas-cloud/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type dashboardSeriesPoint struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type uploadAggregateRow struct {
	Day   string `gorm:"column:day"`
	Total int64  `gorm:"column:total"`
}

func GetDashboardStats(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "User tidak ditemukan")
		return
	}

	var totalFiles int64
	if err := config.DB.Model(&models.File{}).
		Where("user_id = ? AND is_folder = ?", user.ID, false).
		Count(&totalFiles).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal menghitung total file")
		return
	}

	var totalFolders int64
	if err := config.DB.Model(&models.File{}).
		Where("user_id = ? AND is_folder = ?", user.ID, true).
		Count(&totalFolders).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal menghitung total folder")
		return
	}

	var usedStorage int64
	if err := config.DB.Model(&models.File{}).
		Where("user_id = ? AND is_folder = ?", user.ID, false).
		Select("COALESCE(SUM(size), 0)").
		Scan(&usedStorage).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal menghitung storage terpakai")
		return
	}

	var pendingShares int64
	if err := config.DB.Model(&models.FileAccess{}).
		Where("target_email = ? AND status = ?", user.Email, "pending").
		Count(&pendingShares).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal menghitung pending share")
		return
	}

	workingDir, err := os.Getwd()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal membaca direktori kerja")
		return
	}

	volumePath := filepath.VolumeName(workingDir) + "\\"
	if volumePath == "\\" {
		volumePath = workingDir
	}

	totalDisk, freeDisk, err := getDiskUsage(volumePath)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal membaca disk usage")
		return
	}

	uploadSeries, err := buildUploadSeries(uint(user.ID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal membaca traffic upload")
		return
	}

	utils.Success(c, http.StatusOK, "Dashboard stats", gin.H{
		"total_files":      totalFiles,
		"total_folders":    totalFolders,
		"used_storage":     usedStorage,
		"free_storage":     freeDisk,
		"total_disk":       totalDisk,
		"pending_shares":   pendingShares,
		"disk_used":        int64(totalDisk - freeDisk),
		"disk_usage_pct":   calculateUsagePercent(totalDisk, freeDisk),
		"upload_traffic_7": uploadSeries,
	})
}

func buildUploadSeries(userID uint) ([]dashboardSeriesPoint, error) {
	start := time.Now().AddDate(0, 0, -6)
	var rows []uploadAggregateRow

	err := config.DB.
		Table("tb_files").
		Select("DATE(created_at) as day, COALESCE(SUM(size), 0) as total").
		Where("user_id = ? AND is_folder = ? AND created_at >= ?", userID, false, start.Format("2006-01-02 15:04:05")).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	rowMap := make(map[string]int64, len(rows))
	for _, row := range rows {
		rowMap[row.Day] = row.Total
	}

	series := make([]dashboardSeriesPoint, 0, 7)
	for i := 0; i < 7; i++ {
		currentDay := start.AddDate(0, 0, i)
		key := currentDay.Format("2006-01-02")
		series = append(series, dashboardSeriesPoint{
			Label: currentDay.Format("02 Jan"),
			Value: rowMap[key],
		})
	}

	return series, nil
}

func calculateUsagePercent(total, free uint64) float64 {
	if total == 0 {
		return 0
	}

	used := total - free
	return (float64(used) / float64(total)) * 100
}
