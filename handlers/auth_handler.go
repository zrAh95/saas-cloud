package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"saas-cloud/config"
	"saas-cloud/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000 // 6 digit
	return fmt.Sprintf("%06d", otp)
}

func Register(c *gin.Context) {
	var input RegisterInput

	// ambil request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
		})
		return
	}

	// validasi sederhana
	if input.Email == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email & password wajib",
		})
		return
	}

	// hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

	// insert user
	query := "INSERT INTO tb_users (email, password) VALUES (?, ?)"
	result, err := config.DB.Exec(query, input.Email, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal insert user",
		})
		return
	}

	userID, _ := result.LastInsertId()

	// generate OTP
	otpCode := generateOTP()

	// expired 7 menit
	expiredAt := time.Now().UTC().Add(7 * time.Minute)

	// simpan OTP ke DB
	queryOTP := "INSERT INTO tb_otps (user_id, otp_code, expired_at) VALUES (?, ?, ?)"
	_, err = config.DB.Exec(queryOTP, userID, otpCode, expiredAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal simpan OTP",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User berhasil dibuat",
		"user_id": userID,
		"otp":     otpCode,
	})
}

type VerifyOTPInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func VerifyOTP(c *gin.Context) {
	var input VerifyOTPInput

	// ambil request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
		})
		return
	}

	fmt.Println("TIME GO:", time.Now())

	// validasi kosong
	if input.Email == "" || input.OTP == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email & OTP wajib",
		})
		return
	}

	var userID int

	queryUser := "SELECT id FROM tb_users WHERE email = ?"
	err := config.DB.QueryRow(queryUser, input.Email).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
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

	err = config.DB.QueryRow(queryOTP, userID).Scan(&otpCode, &expiredAt, &isUsed)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "OTP tidak ditemukan",
		})
		return
	}

	// validasi kecocokan OTP nya
	if input.OTP != otpCode {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OTP salah",
		})
		return
	}

	// ngecek expired OTP nya
	if time.Now().After(expiredAt) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OTP sudah expired",
		})
		return
	}

	// cek kode OTP kalo udah kepake
	if isUsed == 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OTP sudah digunakan",
		})
		return
	}

	// update user jadi verified
	_, err = config.DB.Exec("UPDATE tb_users SET is_verified = 1 WHERE id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal update user",
		})
		return
	}

	// tandai OTP sudah dipakai
	_, err = config.DB.Exec("UPDATE tb_otps SET is_used = 1 WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal update OTP",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP valid, akun aktif",
	})
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var input LoginInput

	// bind JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
		})
		return
	}

	// validasi kosong
	if input.Email == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email & password wajib",
		})
		return
	}

	var userID int
	var hashedPassword string
	var isVerified int

	// ambil user
	query := "SELECT id, password, is_verified FROM tb_users WHERE email = ?"
	err := config.DB.QueryRow(query, input.Email).Scan(&userID, &hashedPassword, &isVerified)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
		return
	}

	// cek password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password salah",
		})
		return
	}

	// 🔐 generate JWT
	token, err := services.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal generate token",
		})
		return
	}

	// 🔥 response baru
	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   token,
	})
}

func VerifyLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Verify login endpoint hit",
	})
}

func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token kosong"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token salah"})
		return
	}

	tokenString := parts[1]

	// blacklist token
	services.BlacklistToken(tokenString)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout berhasil",
	})
}
