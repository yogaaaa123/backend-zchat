package database

import (
	"fmt"
	"log"
	"time"

	"github.com/satria/obrolan-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	logLevel := logger.Info
	if cfg.AppEnv == "production" {
		logLevel = logger.Warn
	}

	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:                 logger.Default.LogMode(logLevel),
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		})
		if err == nil {
			break
		}
		log.Printf("Database connection attempt %d/5 failed: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}

	// Set connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	DB = db
	log.Println("Database connected successfully")
	return db
}

func GetDB() *gorm.DB {
	return DB
}
