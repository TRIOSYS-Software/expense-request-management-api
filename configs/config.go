package configs

import (
	"os"
	"shwetaik-expense-management-api/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	ServerIP   string
	ServerPort string
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	DB         *gorm.DB
}

func (c *Config) LoadEnv(env string) *Config {
	godotenv.Load(env)

	cfg := &Config{}
	cfg.ServerIP = os.Getenv("SERVER_IP")
	if cfg.ServerIP == "" {
		cfg.ServerIP = "localhost"
	}

	cfg.ServerPort = os.Getenv("SERVER_PORT")
	if cfg.ServerPort == "" {
		cfg.ServerPort = "1234"
	}

	cfg.DBHost = os.Getenv("DB_HOST")
	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}

	cfg.DBUser = os.Getenv("DB_USER")
	if cfg.DBUser == "" {
		cfg.DBUser = "root"
	}

	cfg.DBPassword = os.Getenv("DB_PASSWORD")
	if cfg.DBPassword == "" {
		cfg.DBPassword = "root"
	}

	cfg.DBName = os.Getenv("DB_NAME")
	if cfg.DBName == "" {
		cfg.DBName = "test"
	}

	cfg.DBPort = os.Getenv("DB_PORT")
	if cfg.DBPort == "" {
		cfg.DBPort = "3306"
	}
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
	)
}
