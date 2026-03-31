package main

import (
	"saas-cloud/routes"

	"saas-cloud/config"

	"github.com/gin-gonic/gin"
)

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
