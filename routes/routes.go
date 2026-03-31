package routes

import (
	"saas-cloud/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// AUTH
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/verify", handlers.VerifyOTP)
		auth.POST("/login", handlers.Login)
		auth.POST("/verify-login", handlers.VerifyLogin)
		auth.POST("/logout", handlers.Logout)
	}

	// FILE
	files := api.Group("/files")
	{
		files.POST("/", handlers.UploadFile)
		files.GET("/", handlers.GetFiles)
		files.GET("/:id", handlers.GetFileByID)
		files.DELETE("/:id", handlers.DeleteFile)
		files.POST("/:id/share", handlers.ShareFile)
		files.GET("/shared", handlers.GetSharedFiles)
	}

	api.POST("/files/share/:id/accept", handlers.AcceptShare)

	// USER
	api.GET("/user/profile", handlers.GetProfile)

	// LOG
	api.GET("/logs", handlers.GetLogs)
}
