package configs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	firebase "firebase.google.com/go/v4"
	"gorm.io/gorm"
)

type Config struct {
	ServerIP            string
	ServerPort          string
	DBUser              string
	DBPassword          string
	DBHost              string
	DBPort              string
	DBName              string
	DB                  *gorm.DB
	JWTSecret           string
	SQLACC_API_ENDPOINT string
	SQLACC_API_TOKEN    string
	FRONTEND_URL        string
	EMAIL_USERNAME      string
	EMAIL_PASSWORD      string
	SMTP_HOST           string
	SMTP_PORT           string
	Environment         string
	FirebaseCredPath    string
	FirebaseApp         *firebase.App
	UploadDir           string
}

var requiredEnv = []string{
	"DB_PASSWORD", "JWT_SECRET",
	"SQLACC_API_ENDPOINT", "SQLACC_API_TOKEN",
	"EMAIL_USERNAME", "EMAIL_PASSWORD", "SMTP_HOST",
}

func getEnvOrDefault(env string, defaultValue string) string {
	if value := os.Getenv(env); value != "" {
		return value
	}
	return defaultValue
}

func Validate() error {
	var missing []string
	for _, name := range requiredEnv {
		if os.Getenv(name) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

func loadEnv(env string) *Config {
	godotenv.Load(env)

	cfg := &Config{}
	cfg.ServerIP = getEnvOrDefault("SERVER_IP", "localhost")
	cfg.ServerPort = getEnvOrDefault("SERVER_PORT", "8080")
	cfg.DBHost = getEnvOrDefault("DB_HOST", "localhost")
	cfg.DBUser = getEnvOrDefault("DB_USER", "root")
	cfg.DBPassword = getEnvOrDefault("DB_PASSWORD", "")
	cfg.DBName = getEnvOrDefault("DB_NAME", "test")
	cfg.DBPort = getEnvOrDefault("DB_PORT", "3306")
	cfg.JWTSecret = getEnvOrDefault("JWT_SECRET", "")
	cfg.SQLACC_API_ENDPOINT = getEnvOrDefault("SQLACC_API_ENDPOINT", "")
	cfg.SQLACC_API_TOKEN = getEnvOrDefault("SQLACC_API_TOKEN", "")
	cfg.FRONTEND_URL = getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	cfg.EMAIL_USERNAME = getEnvOrDefault("EMAIL_USERNAME", "")
	cfg.EMAIL_PASSWORD = getEnvOrDefault("EMAIL_PASSWORD", "")
	cfg.SMTP_HOST = getEnvOrDefault("SMTP_HOST", "")
	cfg.SMTP_PORT = getEnvOrDefault("SMTP_PORT", "587")
	cfg.Environment = getEnvOrDefault("ENVIRONMENT", "dev")
	cfg.FirebaseCredPath = getEnvOrDefault("FIREBASE_CREDENTIALS_PATH", "fcm_credentials.json")
	cfg.UploadDir = initUploadDir()

	return cfg
}

func initUploadDir() string {
	workingDir, err := os.Getwd()
	if err == nil {
		uploadDir := filepath.Join(workingDir, "uploads")
		if err := os.MkdirAll(uploadDir, os.ModePerm); err == nil {
			log.Printf("Upload directory initialized at working directory: %s\n", uploadDir)
			return uploadDir
		}
	}

	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		uploadDir := filepath.Join(execDir, "uploads")
		if err := os.MkdirAll(uploadDir, os.ModePerm); err == nil {
			log.Printf("Upload directory initialized at executable location: %s\n", uploadDir)
			return uploadDir
		}
	}

	log.Printf("Warning: Could not determine upload directory, using relative path\n")
	return "uploads"
}

var Envs = loadEnv(".env")
