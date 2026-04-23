package utils

import (
	"regexp"
	"saas-cloud/config"
	"strings"
)

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

// ValidatePassword checks minimum password requirements
func ValidatePassword(password string) map[string]string {
	errors := make(map[string]string)

	if len(password) < 8 {
		errors["length"] = "Password minimal 8 karakter"
	}

	// Cek apakah ada uppercase
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		errors["uppercase"] = "Password harus mengandung huruf besar"
	}

	// Cek apakah ada lowercase
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		errors["lowercase"] = "Password harus mengandung huruf kecil"
	}

	// Cek apakah ada number
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		errors["number"] = "Password harus mengandung angka"
	}

	return errors
}

// CheckEmailExists checks if email already registered
func CheckEmailExists(email string) bool {
	var count int64
	result := config.DB.Table("tb_users").Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false
	}
	return count > 0
}

func NormalizeWhatsAppNumber(phone string) string {
	cleaned := regexp.MustCompile(`[^0-9+]`).ReplaceAllString(strings.TrimSpace(phone), "")
	cleaned = strings.TrimPrefix(cleaned, "+")

	if strings.HasPrefix(cleaned, "0") {
		return "62" + strings.TrimPrefix(cleaned, "0")
	}

	return cleaned
}

func ValidateWhatsAppNumber(phone string) bool {
	normalized := NormalizeWhatsAppNumber(phone)
	return regexp.MustCompile(`^62[0-9]{8,15}$`).MatchString(normalized)
}
