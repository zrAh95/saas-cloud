package services

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var SECRET_KEY []byte

func init() {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY environment variable not set")
	}
	SECRET_KEY = []byte(secretKey)
}

func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": "access",
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SECRET_KEY)
}

func GenerateAccessToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": "access",
		"exp":        time.Now().Add(time.Minute * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SECRET_KEY)
}

func GenerateRefreshToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": "refresh",
		"exp":        time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SECRET_KEY)
}
