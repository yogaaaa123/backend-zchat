package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/database"
	"github.com/satria/obrolan-api/internal/handlers"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/satria/obrolan-api/internal/repositories"
	"github.com/satria/obrolan-api/internal/routes"
	"github.com/satria/obrolan-api/internal/services"
)

// @title           Social Forum API
// @version         1.0.0
// @description     Backend API for social forum with threads, comments, likes, real-time chat, and image uploads.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@socialforum.local

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                Type "Bearer" followed by a space and JWT token.
func main() {
	// Load config
	cfg := config.Load()

	// Connect database
	db := database.Connect(cfg)

	// AutoMigrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Thread{},
		&models.Comment{},
		&models.Like{},
		&models.Message{},
		&models.RefreshToken{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// Init repositories
	userRepo := repositories.NewUserRepository(db)
	threadRepo := repositories.NewThreadRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	likeRepo := repositories.NewLikeRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)

	// Init services
	authService := services.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userService := services.NewUserService(userRepo)

	var imageSvc services.ImageService
	uploadSvc, err := services.NewUploadService(cfg)
	if err != nil {
		log.Printf("Warning: Cloudinary not configured, using no-op upload: %v", err)
		imageSvc = services.NewNoOpUploadService()
		uploadSvc = &services.UploadService{}
	} else {
		imageSvc = uploadSvc
	}

	threadService := services.NewThreadService(threadRepo, likeRepo, imageSvc)
	commentService := services.NewCommentService(commentRepo)
	likeService := services.NewLikeService(likeRepo, threadRepo)

	// Init WebSocket Hub
	hub := services.NewHub()
	go hub.Run()

	// Init handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	threadHandler := handlers.NewThreadHandler(threadService)
	commentHandler := handlers.NewCommentHandler(commentService)
	likeHandler := handlers.NewLikeHandler(likeService)
	uploadHandler := handlers.NewUploadHandler(uploadSvc)
	chatHandler := handlers.NewChatHandler(hub, messageRepo, cfg)

	// Periodic cleanup expired tokens
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			refreshTokenRepo.DeleteExpired()
		}
	}()

	// Init Gin router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	if cfg.AppEnv != "production" {
		r.Use(gin.Logger())
	}
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup routes
	routes.Setup(r, cfg, &routes.Handlers{
		Auth:    authHandler,
		User:    userHandler,
		Thread:  threadHandler,
		Comment: commentHandler,
		Like:    likeHandler,
		Upload:  uploadHandler,
		Chat:    chatHandler,
	})

	// Start server
	addr := ":" + cfg.AppPort
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
