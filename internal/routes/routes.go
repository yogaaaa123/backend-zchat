package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/handlers"
	"github.com/satria/obrolan-api/internal/middleware"

	// Swagger docs (blank import agar init() berjalan)
	_ "github.com/satria/obrolan-api/docs"
)

type Handlers struct {
	Auth    *handlers.AuthHandler
	User    *handlers.UserHandler
	Thread  *handlers.ThreadHandler
	Comment *handlers.CommentHandler
	Like    *handlers.LikeHandler
	Upload  *handlers.UploadHandler
	Chat    *handlers.ChatHandler
}

func Setup(r *gin.Engine, cfg *config.Config, h *Handlers) {
	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Social Forum API is running",
		})
	})

	// ==================== Public Routes ====================
	auth := api.Group("/auth")
	auth.Use(middleware.RateLimitMiddleware(20, 1*time.Minute)) // 20 req/min per IP
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
	}

	// ==================== Protected Routes (JWT required) ====================
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// Auth (protected)
		protected.GET("/auth/me", h.Auth.Me)
		protected.POST("/auth/logout", h.Auth.Logout)

		// Users
		protected.GET("/users/me", h.User.GetProfile)
		protected.PUT("/users/me", h.User.UpdateProfile)
		protected.PUT("/users/me/password", h.User.ChangePassword)
		protected.GET("/users/:id", h.User.GetPublicProfile)

		// Upload
		protected.POST("/upload", h.Upload.Upload)

		// Threads
		protected.GET("/threads", h.Thread.GetAll)
		protected.GET("/threads/search", h.Thread.Search)
		protected.POST("/threads", h.Thread.Create)
		protected.GET("/threads/:id", h.Thread.GetByID)
		protected.PUT("/threads/:id", h.Thread.Update)
		protected.DELETE("/threads/:id", h.Thread.Delete)

		// Comments (nested under threads)
		protected.POST("/threads/:id/comments", h.Comment.Create)
		protected.GET("/threads/:id/comments", h.Comment.GetByThread)
		protected.DELETE("/comments/:id", h.Comment.Delete)

		// Likes
		protected.POST("/threads/:id/like", h.Like.Toggle)

		// Chat messages
		protected.GET("/threads/:id/messages", h.Chat.GetMessages)
	}

	// ==================== WebSocket (auth via query param) ====================
	r.GET("/api/v1/threads/:id/chat", h.Chat.HandleWebSocket)
}
