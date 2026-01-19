package configs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	helper "shwetaik-expense-management-api/Helper"
	"shwetaik-expense-management-api/models"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

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
	EMAIL_USERNAME      string
	EMAIL_PASSWORD      string
	SMTP_HOST           string
	SMTP_PORT           string
	Environment         string
	FirebaseCredPath    string
	FirebaseApp         *firebase.App
	UploadDir           string
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

func (c *Config) ConnectDB() error {
	dsn := c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	c.DB = db
	return nil
}

func (c *Config) SetupFirebase() error {
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		credPath = filepath.Join(".", "fcm_credentials.json")
	}
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		log.Printf("Firebase credentials file does not exist at path: %s\n", credPath)
		return err
	}
	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Printf("error initializing Firebase app: %v\n", err)
		return err
	}
	log.Printf("Firebase app initialized: %+v\n", app)
	log.Printf("Option data: %+v\n", opt)
	c.FirebaseApp = app
	return nil
}

func (c *Config) InitializedDB() {
	c.DB.AutoMigrate(
		&models.Users{},
		&models.ExpenseRequests{},
		&models.ExpenseRequestAttachments{},
		&models.ExpenseItems{},
		&models.ApprovalPolicies{},
		&models.ExpenseApprovals{},
		&models.Roles{},
		&models.Permissions{},
		&models.Departments{},
		&models.ExpenseCategories{},
		&models.ApprovalPoliciesUsers{},
		&models.PaymentMethod{},
		&models.Project{},
		&models.GLAcc{},
		&models.PasswordReset{},
		&models.Notification{},
		&models.DeviceToken{},
	)

	if err := SeedPermissions(c.DB); err != nil {
		log.Fatalf("Failed to seed permissions: %v", err)
	}

	roles := []models.Roles{
		{Name: "Admin", Description: "Admin with full privileges", IsAdmin: true},
	}

	for _, role := range roles {
		var existing models.Roles
		if err := c.DB.Where("name = ?", role.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			c.DB.Create(&role)
			fmt.Printf("✅ Seeded role: %s\n", role.Name)
		}
	}

	var adminRole models.Roles
	c.DB.Where("name = ?", "Admin").First(&adminRole)

	var allPerms []models.Permissions
	c.DB.Not("entity = ? AND action = ?", "expense-request", "create").Find(&allPerms)

	if err := c.DB.Model(&adminRole).Association("Permissions").Replace(allPerms); err != nil {
		log.Fatalf("Failed to assign permissions to Admin: %v", err)
	}

	fmt.Println("✅ All permissions assigned to Admin role")

	var count int64
	c.DB.Model(&models.Users{}).Where("email = ?", "admin@example.com").Count(&count)
	if count == 0 {
		adminUser := models.Users{
			Name:   "Admin",
			Email:  "admin@example.com",
			RoleID: adminRole.ID,
		}

		hashPassword, err := helper.HashPassword("admin")
		if err != nil {
			panic(err)
		}
		adminUser.Password = hashPassword
		c.DB.Create(&adminUser)

		fmt.Println("✅ Default admin user created")
	}
}

func SeedPermissions(db *gorm.DB) error {
	permissions := []models.Permissions{
		// Dashboard
		{Name: "Dashboard", Entity: "dashboard", Action: "view", ActionName: "View Dashboard"},

		// Expense Request
		{Name: "Expense Request", Entity: "expense-request", Action: "view", ActionName: "View Expense Requests"},
		{Name: "Expense Request", Entity: "expense-request", Action: "create", ActionName: "Create Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "update", ActionName: "Update Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "delete", ActionName: "Delete Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "approve", ActionName: "Approve Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "reject", ActionName: "Reject Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "send-to-sqlacc", ActionName: "Send To SQL Account"},
		{Name: "Expense Request", Entity: "expense-request", Action: "export", ActionName: "Export Expense Requests"},

		// User
		{Name: "User", Entity: "user", Action: "view", ActionName: "View Users"},
		{Name: "User", Entity: "user", Action: "create", ActionName: "Create User"},
		{Name: "User", Entity: "user", Action: "update", ActionName: "Update User"},
		{Name: "User", Entity: "user", Action: "delete", ActionName: "Delete User"},
		{Name: "User", Entity: "user", Action: "export", ActionName: "Export Users"},

		// Roles
		{Name: "Role", Entity: "role", Action: "view", ActionName: "View Roles"},
		{Name: "Role", Entity: "role", Action: "create", ActionName: "Create Role"},
		{Name: "Role", Entity: "role", Action: "update", ActionName: "Update Role"},
		{Name: "Role", Entity: "role", Action: "delete", ActionName: "Delete Role"},
		{Name: "Role", Entity: "role", Action: "export", ActionName: "Export Roles"},

		// Departments
		{Name: "Department", Entity: "department", Action: "view", ActionName: "View Departments"},
		{Name: "Department", Entity: "department", Action: "create", ActionName: "Create Department"},
		{Name: "Department", Entity: "department", Action: "update", ActionName: "Update Department"},
		{Name: "Department", Entity: "department", Action: "delete", ActionName: "Delete Department"},
		{Name: "Department", Entity: "department", Action: "export", ActionName: "Export Departments"},

		// Policies
		{Name: "Policy", Entity: "policy", Action: "view", ActionName: "View Policies"},
		{Name: "Policy", Entity: "policy", Action: "create", ActionName: "Create Policy"},
		{Name: "Policy", Entity: "policy", Action: "update", ActionName: "Update Policy"},
		{Name: "Policy", Entity: "policy", Action: "delete", ActionName: "Delete Policy"},
		{Name: "Policy", Entity: "policy", Action: "export", ActionName: "Export Policies"},

		// GL Accounts
		{Name: "GL Account", Entity: "gl-account", Action: "view-gl-accounts", ActionName: "View GL Accounts"},
		{Name: "GL Account", Entity: "gl-account", Action: "sync-gl-accounts", ActionName: "Sync GL Accounts"},
		{Name: "GL Account", Entity: "gl-account", Action: "export-gl-accounts", ActionName: "Export GL Accounts"},
		{Name: "GL Account", Entity: "gl-account", Action: "view-assigned-gl-accounts", ActionName: "View Assigned GL Accounts"},
		{Name: "GL Account", Entity: "gl-account", Action: "create-assign-gl-account", ActionName: "Create Assigned GL Account"},
		{Name: "GL Account", Entity: "gl-account", Action: "edit-assigned-gl-account", ActionName: "Edit Assigned GL Account"},
		{Name: "GL Account", Entity: "gl-account", Action: "delete-assigned-gl-account", ActionName: "Delete Assigned GL Account"},
		{Name: "GL Account", Entity: "gl-account", Action: "export-assigned-gl-accounts", ActionName: "Export Assigned GL Accounts"},

		// Payment Methods
		{Name: "Payment Method", Entity: "payment-method", Action: "view-payment-methods", ActionName: "View Payment Methods"},
		{Name: "Payment Method", Entity: "payment-method", Action: "sync-payment-methods", ActionName: "Sync Payment Methods"},
		{Name: "Payment Method", Entity: "payment-method", Action: "export-payment-methods", ActionName: "Export Payment Methods"},
		{Name: "Payment Method", Entity: "payment-method", Action: "view-assigned-payment-methods", ActionName: "View Assigned Payment Methods"},
		{Name: "Payment Method", Entity: "payment-method", Action: "create-assign-payment-method", ActionName: "Create Assigned Payment Method"},
		{Name: "Payment Method", Entity: "payment-method", Action: "edit-assigned-payment-method", ActionName: "Edit Assigned Payment Method"},
		{Name: "Payment Method", Entity: "payment-method", Action: "delete-assigned-payment-method", ActionName: "Delete Assigned Payment Method"},
		{Name: "Payment Method", Entity: "payment-method", Action: "export-assigned-payment-methods", ActionName: "Export Assigned Payment Methods"},

		// Projects
		{Name: "Project", Entity: "project", Action: "view-projects", ActionName: "View Projects"},
		{Name: "Project", Entity: "project", Action: "sync-projects", ActionName: "Sync Projects"},
		{Name: "Project", Entity: "project", Action: "export-projects", ActionName: "Export Projects"},
		{Name: "Project", Entity: "project", Action: "view-assigned-projects", ActionName: "View Assigned Projects"},
		{Name: "Project", Entity: "project", Action: "create-assign-project", ActionName: "Create Assigned Project"},
		{Name: "Project", Entity: "project", Action: "edit-assigned-project", ActionName: "Edit Assigned Project"},
		{Name: "Project", Entity: "project", Action: "delete-assigned-project", ActionName: "Delete Assigned Project"},
		{Name: "Project", Entity: "project", Action: "export-assigned-projects", ActionName: "Export Assigned Projects"},
	}

	for _, perm := range permissions {
		var existing models.Permissions
		err := db.Where("entity = ? AND action = ?", perm.Entity, perm.Action).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("failed to seed permission %s:%s: %v", perm.Entity, perm.Action, err)
			}
			fmt.Printf("✅ Seeded permission: %s:%s\n", perm.Entity, perm.Action)
		}
	}

	fmt.Println("✅ All permissions seeded successfully.")
	return nil
}

var Envs = loadEnv(".env")
