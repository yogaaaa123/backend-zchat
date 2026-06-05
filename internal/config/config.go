package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort  string
	AppEnv   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret              string
	JWTExpiresIn           time.Duration
	RefreshTokenExpiresIn  time.Duration

	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string

	CORSOrigins []string
}

func Load() *Config {
	godotenv.Load() // ignore error if .env not found

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "social_forum"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:              getJWTSecret(),
		JWTExpiresIn:           getDuration("JWT_EXPIRES_IN", 72*time.Hour),
		RefreshTokenExpiresIn:  getDuration("REFRESH_TOKEN_EXPIRES_IN", 168*time.Hour),

		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),

		CORSOrigins: getSlice("CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
	}
}

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		if os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging" {
			log.Fatal("FATAL: JWT_SECRET must be set in " + os.Getenv("APP_ENV") + " environment")
		}
		return "dev-secret-not-for-production"
	}
	return secret
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getSlice(key string, fallback []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return strings.Split(val, ",")
}

func getDuration(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}
