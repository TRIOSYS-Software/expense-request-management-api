package configs

import (
	"os"
	helper "shwetaik-expense-management-api/Helper"
	"shwetaik-expense-management-api/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
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
	SQLACC_API_PASSWORD string
	SQLACC_API_KEY      string
	SQLACC_API_URL      string
	FILTER_GL_CODES     string
	FRONTEND_URL        string
}

func getEnvOrDefault(env string, defaultValue string) string {
	value := os.Getenv(env)
	if value == "" {
		if defaultValue != "" {
			return defaultValue
		} else {
			panic("Environment variable " + env + " is not set")
		}
	}
	return value
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
	cfg.SQLACC_API_PASSWORD = getEnvOrDefault("SQLACC_API_PASSWORD", "")
	cfg.SQLACC_API_KEY = getEnvOrDefault("SQLACC_API_KEY", "")
	cfg.SQLACC_API_URL = getEnvOrDefault("SQLACC_API_URL", "")
	cfg.FILTER_GL_CODES = getEnvOrDefault("FILTER_GL_CODES", "")
	cfg.FRONTEND_URL = getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	return cfg
}

func (c *Config) ConnectDB() error {
	dsn := c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	c.DB = db
	return nil
}

func (c *Config) InitializedDB() {
	c.DB.AutoMigrate(
		&models.Users{},
		&models.ExpenseRequests{},
		&models.ExpenseItems{},
		&models.ApprovalPolicies{},
		&models.ExpenseApprovals{},
		&models.Roles{},
		&models.Departments{},
		&models.ExpenseCategories{},
		&models.ApprovalPoliciesUsers{},
		&models.PaymentMethod{},
		&models.Project{},
		&models.GLAcc{},
		&models.PasswordReset{},
	)

	var role models.Roles
	c.DB.First(&role, "name = ?", "Admin")
	if role.ID == 0 {
		roles := []models.Roles{
			{Name: "Admin"},
			{Name: "Approver"},
			{Name: "Staff"},
		}
		c.DB.Create(&roles)
	}

	var count int64
	c.DB.Model(&models.Users{}).Where("name = ?", "Admin").Count(&count)
	if count == 0 {
		adminUser := models.Users{
			Name:     "Admin",
			Email:    "admin@example.com",
			Password: "admin",
			RoleID:   1,
		}

		hashPassword, err := helper.HashPassword(adminUser.Password)
		if err != nil {
			panic(err)
		}
		adminUser.Password = hashPassword

		c.DB.Create(&adminUser)
	}
}

var Envs = loadEnv(".env")
