package testutil

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"shwetaik-expense-management-api/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func allModels() []any {
	return []any{
		&models.Users{},
		&models.ExpenseRequests{},
		&models.ExpenseRequestAttachments{},
		&models.ApprovalPolicies{},
		&models.ExpenseApprovals{},
		&models.AdvanceRequests{},
		&models.AdvanceRequestAttachments{},
		&models.AdvanceApprovals{},
		&models.Roles{},
		&models.Permissions{},
		&models.Departments{},
		&models.ExpenseCategories{},
		&models.ApprovalPoliciesUsers{},
		&models.ApprovalPolicyGLAccount{},
		&models.PaymentMethod{},
		&models.Project{},
		&models.GLAcc{},
		&models.PasswordReset{},
		&models.Notification{},
		&models.DeviceToken{},
	}
}

var (
	sharedDB   *gorm.DB
	sharedErr  error
	sharedOnce sync.Once
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func dsn() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		env("TEST_DB_USER", "root"),
		env("TEST_DB_PASSWORD", "testpassword"),
		env("TEST_DB_HOST", "127.0.0.1"),
		env("TEST_DB_PORT", "3307"),
		env("TEST_DB_NAME", "expense_test"),
	)
}

func connect() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to test database: %w", err)
	}
	if err := db.AutoMigrate(allModels()...); err != nil {
		return nil, fmt.Errorf("migrate test schema: %w", err)
	}
	return db, nil
}

func DB(t *testing.T) *gorm.DB {
	t.Helper()

	sharedOnce.Do(func() { sharedDB, sharedErr = connect() })
	if sharedErr != nil {
		t.Fatalf("test database unavailable (run `make test-up`): %v", sharedErr)
	}

	Reset(t, sharedDB)
	return sharedDB
}

func Reset(t *testing.T, db *gorm.DB) {
	t.Helper()

	var tables []string
	if err := db.Raw(`
		SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
	`).Scan(&tables).Error; err != nil {
		t.Fatalf("list tables for reset: %v", err)
	}

	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		t.Fatalf("disable foreign key checks: %v", err)
	}
	defer func() {
		if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
			t.Fatalf("re-enable foreign key checks: %v", err)
		}
	}()

	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE `" + table + "`").Error; err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}
