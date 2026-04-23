package routes

import (
	"saas-cloud/handlers"
	"saas-cloud/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/verify", handlers.VerifyOTP)
		auth.POST("/login", handlers.Login)
		auth.POST("/logout", handlers.Logout)
		auth.POST("/refresh", handlers.RefreshToken)
	}

	authProtected := api.Group("/auth")
	authProtected.Use(middleware.AuthMiddleware())
	{
		authProtected.GET("/profile", handlers.GetProfile)
		authProtected.PUT("/profile", handlers.UpdateProfile)
		authProtected.GET("/profile/avatar", handlers.GetProfileAvatar)
		authProtected.POST("/profile/avatar", handlers.UploadProfileAvatar)
	}

	dashboard := api.Group("/dashboard")
	dashboard.Use(middleware.AuthMiddleware())
	{
		dashboard.GET("/stats", handlers.GetDashboardStats)
	}

	files := api.Group("/files")
	files.Use(middleware.AuthMiddleware())
	{
		files.POST("/folders", handlers.CreateFolder)
		files.POST("/upload", handlers.UploadFile)
		files.GET("/", handlers.GetFiles)
		files.GET("/folders/:id/contents", handlers.GetFolderContents)
		files.GET("/shared", handlers.GetSharedFiles)
		files.POST("/shares/:id/accept", handlers.AcceptShare)
		files.GET("/:id", handlers.GetFileByID)
		files.GET("/:id/download", handlers.DownloadFile)
		files.DELETE("/:id", handlers.DeleteFile)
		files.POST("/:id/share", handlers.ShareFile)
		files.DELETE("/:id/share", handlers.RevokeFileAccess)
	}

	logs := api.Group("/logs")
	logs.Use(middleware.AuthMiddleware())
	{
		logs.GET("/", handlers.GetLogs)
	}
}
