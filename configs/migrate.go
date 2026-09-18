package configs

import (
	"fmt"
	"log"
	"strings"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/security"

	"gorm.io/gorm"
)

func (c *Config) InitializedDB() {
	c.DB.AutoMigrate(
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
	)

	c.DB.Exec("ALTER TABLE approval_policies DROP COLUMN IF EXISTS gl_account_id")

	// Backfill: legacy approval_policies rows have NULL/empty policy_type — pin to 'expense'.
	c.DB.Exec("UPDATE approval_policies SET policy_type = 'expense' WHERE policy_type IS NULL OR policy_type = ''")

	// Extend advance_requests.status enum with 'closed' if the column is still on the old
	// value set
	var advanceStatusType string
	c.DB.Raw(`
		SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'advance_requests'
		  AND COLUMN_NAME = 'status'
	`).Scan(&advanceStatusType)
	if advanceStatusType != "" && !strings.Contains(advanceStatusType, "'closed'") {
		if err := c.DB.Exec(`
			ALTER TABLE advance_requests
			MODIFY COLUMN status ENUM('pending','approved','rejected','completed','closed')
			NOT NULL DEFAULT 'pending'
		`).Error; err != nil {
			log.Printf("Failed to extend advance_requests.status enum: %v", err)
		} else {
			fmt.Println("✅ advance_requests.status enum extended with 'closed'")
		}
	}

	var expenseStatusType string
	c.DB.Raw(`
		SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'expense_requests'
		  AND COLUMN_NAME = 'status'
	`).Scan(&expenseStatusType)
	if expenseStatusType != "" && !strings.Contains(expenseStatusType, "'completed'") {
		if err := c.DB.Exec(`
			ALTER TABLE expense_requests
			MODIFY COLUMN status ENUM('pending','approved','rejected','completed')
			NOT NULL DEFAULT 'pending'
		`).Error; err != nil {
			log.Printf("Failed to extend expense_requests.status enum: %v", err)
		} else {
			fmt.Println("✅ expense_requests.status enum extended with 'completed'")
		}
	}

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
	c.DB.Where("NOT (entity IN ? AND action = ?)", []string{"expense-request", "advance-request"}, "create").Find(&allPerms)

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

		hashPassword, err := security.HashPassword("admin")
		if err != nil {
			panic(err)
		}
		adminUser.Password = hashPassword
		c.DB.Create(&adminUser)

		fmt.Println("✅ Default admin user created")
	}
}
