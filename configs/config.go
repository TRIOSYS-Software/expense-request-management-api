package configs

import (
	"errors"
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
	FILTER_GL_CODE      string
}

func loadEnv(env string) *Config {
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

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		panic(errors.New("JWT_SECRET is not set"))
	}

	cfg.SQLACC_API_PASSWORD = os.Getenv("SQLACC_API_PASSWORD")
	if cfg.SQLACC_API_PASSWORD == "" {
		panic(errors.New("SQLACC_API_PASSWORD is not set"))
	}

	cfg.SQLACC_API_KEY = os.Getenv("SQLACC_API_KEY")
	if cfg.SQLACC_API_KEY == "" {
		panic(errors.New("SQLACC_API_KEY is not set"))
	}
	cfg.SQLACC_API_URL = os.Getenv("SQLACC_API_URL")
	if cfg.SQLACC_API_URL == "" {
		panic(errors.New("SQLACC_API_URL is not set"))
	}

	cfg.FILTER_GL_CODE = os.Getenv("FILTER_GL_CODE")
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
