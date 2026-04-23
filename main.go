package main

import (
	"log"
	"net/http"

	"saas-cloud/config"
	"saas-cloud/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
}

func main() {
	config.ConnectDB()

	r := gin.Default()

	// Frontend monolith: asset Mazer hasil build tetap diserve dari app Go yang sama.
	r.Static("/mazer-assets", "./web/assets/dist/assets")
	r.Static("/frontend-js", "./web/js")
	r.Static("/frontend-css", "./web/css")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/login", func(c *gin.Context) {
		c.File("./web/pages/login.html")
	})

	r.GET("/register", func(c *gin.Context) {
		c.File("./web/pages/register.html")
	})

	r.GET("/dashboard", func(c *gin.Context) {
		c.File("./web/pages/dashboard.html")
	})

	r.GET("/profile-app", func(c *gin.Context) {
		c.File("./web/pages/profile.html")
	})

	r.GET("/files-app", func(c *gin.Context) {
		c.File("./web/pages/files.html")
	})

	r.GET("/shared-app", func(c *gin.Context) {
		c.File("./web/pages/shared.html")
	})

	r.GET("/folder-app", func(c *gin.Context) {
		c.File("./web/pages/folder.html")
	})

	routes.SetupRoutes(r)

	r.Run(":8080")
}
