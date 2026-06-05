package services

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/database"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/satria/obrolan-api/internal/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testCfg    *config.Config
	testUserRepo      *repositories.UserRepository
	testRefreshRepo   *repositories.RefreshTokenRepository
	testThreadRepo    *repositories.ThreadRepository
	testCommentRepo   *repositories.CommentRepository
	testLikeRepo      *repositories.LikeRepository
	testAuthSvc       *AuthService
	testUserSvc       *UserService
	testThreadSvc     *ThreadService
	testCommentSvc    *CommentService
	testLikeSvc       *LikeService
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	// In-memory SQLite
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect test DB: %v", err)
	}

	// Override global DB
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	database.DB = db

	// AutoMigrate
	if err := db.AutoMigrate(
		&models.User{},
		&models.Thread{},
		&models.Comment{},
		&models.Like{},
		&models.Message{},
		&models.RefreshToken{},
	); err != nil {
		log.Fatalf("Failed to migrate test DB: %v", err)
	}

	// Test config
	testCfg = &config.Config{
		AppPort:              "8080",
		AppEnv:               "test",
		JWTSecret:            "test-secret-key-for-jwt",
		JWTExpiresIn:         15 * time.Minute,
		RefreshTokenExpiresIn: 168 * time.Hour,
	}

	// Init repos
	testUserRepo = repositories.NewUserRepository(db)
	testRefreshRepo = repositories.NewRefreshTokenRepository(db)
	testThreadRepo = repositories.NewThreadRepository(db)
	testCommentRepo = repositories.NewCommentRepository(db)
	testLikeRepo = repositories.NewLikeRepository(db)

	// Init services
	testAuthSvc = NewAuthService(testUserRepo, testRefreshRepo, testCfg)
	testUserSvc = NewUserService(testUserRepo)
	testThreadSvc = NewThreadService(testThreadRepo, testLikeRepo, NewNoOpUploadService())
	testCommentSvc = NewCommentService(testCommentRepo)
	testLikeSvc = NewLikeService(testLikeRepo, testThreadRepo)
}

func cleanDB(t *testing.T) {
	t.Helper()
	database.GetDB().Exec("DELETE FROM likes")
	database.GetDB().Exec("DELETE FROM comments")
	database.GetDB().Exec("DELETE FROM threads")
	database.GetDB().Exec("DELETE FROM refresh_tokens")
	database.GetDB().Exec("DELETE FROM users")
}

func seedUser(t *testing.T) (userID string) {
	t.Helper()

	id := uuid.New().String()[:8]
	req := RegisterRequest{
		Username: "testuser_" + id,
		Email:    "test_" + id + "@test.com",
		Password: "password123",
	}

	resp, err := testAuthSvc.Register(req)
	if err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	return resp.User.ID.String()
}

func seedThread(t *testing.T, userID string) string {
	t.Helper()

	uid, _ := uuid.Parse(userID)
	req := CreateThreadRequest{
		Title:   "Test Thread Title " + uuid.New().String()[:8],
		Content: "This is the content of the test thread for testing purposes. " + uuid.New().String(),
	}

	resp, err := testThreadSvc.Create(uid, req)
	if err != nil {
		t.Fatalf("Failed to seed thread: %v", err)
	}

	return resp.ID.String()
}
