package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"saas-cloud/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token kosong"})
			c.Abort()
			return
		}

		// split Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token salah"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// cek blacklist
		if services.IsTokenBlacklisted(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token sudah logout"})
			c.Abort()
			return
		}

		// parse token dengan algorithm validation
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return services.SECRET_KEY, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
			c.Abort()
			return
		}

		// ambil claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal membaca token"})
			c.Abort()
			return
		}

		// ambil user_id dengan safe type assertion
		userIDFloat, exists := claims["user_id"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
			c.Abort()
			return
		}

		userID, ok := userIDFloat.(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
			c.Abort()
			return
		}

		tokenType, ok := claims["token_type"].(string)
		if !ok || tokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token tidak valid"})
			c.Abort()
			return
		}

		// simpan ke context
		c.Set("user_id", int(userID))

		c.Next()
	}
}
