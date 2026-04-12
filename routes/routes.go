package routes

import (
	"saas-cloud/handlers"
	"saas-cloud/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// PUBLIC
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/verify", handlers.VerifyOTP)
		auth.POST("/login", handlers.Login)
		auth.POST("/logout", handlers.Logout)
		//buat refresh token 
		auth.POST("/refresh", handlers.RefreshToken)
	}

	// 🔐 PROTECTED
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.AuthMiddleware())
	{
		authProtected.GET("/profile", handlers.GetProfile)
	}

	// FILE
	files := api.Group("/files")
	files.Use(middleware.AuthMiddleware())
	{
		files.POST("/upload", handlers.UploadFile)
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

