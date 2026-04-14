package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"saas-cloud/config"
	"saas-cloud/services"
	"saas-cloud/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000
	return fmt.Sprintf("%06d", otp)
}

func Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid input")
		return
	}

	// Trim whitespace
	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)

	if input.Email == "" || input.Password == "" {
		utils.Error(c, http.StatusBadRequest, "Email & password wajib")
		return
	}

	// Validate email format
	if !utils.ValidateEmail(input.Email) {
		utils.Error(c, http.StatusBadRequest, "Format email tidak valid")
		return
	}

	// Check email already exists
	if utils.CheckEmailExists(input.Email) {
		utils.Error(c, http.StatusBadRequest, "Email sudah terdaftar")
		return
	}

	// Validate password strength
	passwordErrors := utils.ValidatePassword(input.Password)
	if len(passwordErrors) > 0 {
		utils.Error(c, http.StatusBadRequest, "Password tidak memenuhi kriteria: "+passwordErrors["length"])
		return
	}

	// Rate limiting: 5 attempts per 15 minutes
	if !services.CheckRateLimit("register:"+input.Email, 5, 900) {
		utils.Error(c, http.StatusTooManyRequests, "Terlalu banyak upaya registrasi, coba lagi dalam beberapa menit")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	query := "INSERT INTO tb_users (email, password) VALUES (?, ?)"
	result, err := config.SQLDB.Exec(query, input.Email, string(hashedPassword))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	userID, _ := result.LastInsertId()

	otpCode := generateOTP()
	expiredAt := time.Now().UTC().Add(7 * time.Minute)

	queryOTP := "INSERT INTO tb_otps (user_id, otp_code, expired_at) VALUES (?, ?, ?)"
	_, err = config.SQLDB.Exec(queryOTP, userID, otpCode, expiredAt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	// Log: In production, log this OTP ke console/email, tapi buat dev gw cout ke log
	fmt.Printf("[OTP] Email: %s, OTP: %s\n", input.Email, otpCode)

	services.ResetRateLimit("register:" + input.Email)
	utils.Success(c, http.StatusOK, "Registrasi berhasil, silakan cek OTP", nil)
}

type VerifyOTPInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func VerifyOTP(c *gin.Context) {
	var input VerifyOTPInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid input")
		return
	}

	if input.Email == "" || input.OTP == "" {
		utils.Error(c, http.StatusBadRequest, "Email & OTP wajib")
		return
	}

	// Rate limiting: max 10 failed OTP attempts per 15 minutes
	if !services.CheckRateLimit("verify_otp:"+input.Email, 10, 900) {
		utils.Error(c, http.StatusTooManyRequests, "Terlalu banyak percobaan verifikasi OTP. Coba lagi dalam 15 menit")
		return
	}

	var userID int

	queryUser := "SELECT id FROM tb_users WHERE email = ?"
	err := config.SQLDB.QueryRow(queryUser, input.Email).Scan(&userID)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "OTP tidak valid atau expired")
		return
	}

	var otpCode string
	var expiredAt time.Time
	var isUsed int

	queryOTP := `
	SELECT otp_code, expired_at, is_used 
	FROM tb_otps 
	WHERE user_id = ? 
	ORDER BY id DESC 
	LIMIT 1`

	err = config.SQLDB.QueryRow(queryOTP, userID).Scan(&otpCode, &expiredAt, &isUsed)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "OTP tidak valid atau expired")
		return
	}

	if input.OTP != otpCode || time.Now().After(expiredAt) || isUsed == 1 {
		utils.Error(c, http.StatusBadRequest, "OTP tidak valid atau expired")
		return
	}

	_, err = config.SQLDB.Exec("UPDATE tb_users SET is_verified = 1 WHERE id = ?", userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	_, err = config.SQLDB.Exec("UPDATE tb_otps SET is_used = 1 WHERE user_id = ?", userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	// Reset rate limit on successful verification
	services.ResetRateLimit("verify_otp:" + input.Email)

	utils.Success(c, http.StatusOK, "Verifikasi berhasil", nil)
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid input")
		return
	}

	// Trim whitespace
	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)

	if input.Email == "" || input.Password == "" {
		utils.Error(c, http.StatusBadRequest, "Email & password wajib")
		return
	}

	// Validate email format
	if !utils.ValidateEmail(input.Email) {
		utils.Error(c, http.StatusBadRequest, "Format email tidak valid")
		return
	}

	// Rate limiting: max 5 failed attempts per 15 minutes
	if !services.CheckRateLimit("login:"+input.Email, 5, 900) {
		remaining := services.GetRateLimitRemaining("login:"+input.Email, 5, 900)
		utils.Error(c, http.StatusTooManyRequests, 
			fmt.Sprintf("Terlalu banyak percobaan login. Coba lagi dalam 15 menit. Sisa percobaan: %d", remaining))
		return
	}

	var userID int
	var hashedPassword string
	var isVerified int

	query := "SELECT id, password, is_verified FROM tb_users WHERE email = ?"
	err := config.SQLDB.QueryRow(query, input.Email).Scan(&userID, &hashedPassword, &isVerified)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	if isVerified == 0 {
		utils.Error(c, http.StatusUnauthorized, "Akun belum terverifikasi")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	accessToken, err := services.GenerateAccessToken(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	refreshToken, err := services.GenerateRefreshToken(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	// Reset rate limit on successful login
	services.ResetRateLimit("login:" + input.Email)

	utils.Success(c, http.StatusOK, "Login berhasil", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid input")
		return
	}

	token, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return services.SECRET_KEY, nil
	})

	if err != nil || !token.Valid {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	// Safe type assertion with error handling
	userIDFloat, exists := claims["user_id"]
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	userID, ok := userIDFloat.(float64)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	newAccessToken, err := services.GenerateAccessToken(int(userID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	utils.Success(c, http.StatusOK, "Refresh token berhasil", gin.H{
		"access_token": newAccessToken,
	})
}

func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	tokenString := parts[1]

	services.BlacklistToken(tokenString)

	utils.Success(c, http.StatusOK, "Logout Berhasil", nil)
}