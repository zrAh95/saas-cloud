package main

import (
	"log"

	"saas-cloud/config"
	"saas-cloud/routes"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
}

func main() {
	config.ConnectDB()

	r := gin.Default()

	// 🔥 DEBUG ROUTE (TARO DI SINI)
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "ok"})
	})

	routes.SetupRoutes(r)

	r.Run(":8080")
}
