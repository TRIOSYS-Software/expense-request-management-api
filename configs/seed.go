package configs

import (
	"fmt"

	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

func SeedPermissions(db *gorm.DB) error {
	permissions := []models.Permissions{
		// Dashboard
		{Name: "Dashboard", Entity: "dashboard", Action: "view", ActionName: "View Dashboard"},

		// Expense Request
		{Name: "Expense Request", Entity: "expense-request", Action: "view", ActionName: "View Expense Requests"},
		{Name: "Expense Request", Entity: "expense-request", Action: "create", ActionName: "Create Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "update", ActionName: "Update Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "delete", ActionName: "Delete Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "soft-delete", ActionName: "Archive Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "approve", ActionName: "Approve Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "reject", ActionName: "Reject Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "send-to-sqlacc", ActionName: "Send To SQL Account"},
		{Name: "Expense Request", Entity: "expense-request", Action: "complete", ActionName: "Manually Complete Expense Request"},
		{Name: "Expense Request", Entity: "expense-request", Action: "export", ActionName: "Export Expense Requests"},

		// Advance Request
		{Name: "Advance Request", Entity: "advance-request", Action: "view", ActionName: "View Advance Requests"},
		{Name: "Advance Request", Entity: "advance-request", Action: "create", ActionName: "Create Advance Request"},
		{Name: "Advance Request", Entity: "advance-request", Action: "update", ActionName: "Update Advance Request"},
		{Name: "Advance Request", Entity: "advance-request", Action: "delete", ActionName: "Delete Advance Request"},
		{Name: "Advance Request", Entity: "advance-request", Action: "soft-delete", ActionName: "Archive Advance Request"},
		{Name: "Advance Request", Entity: "advance-request", Action: "approve", ActionName: "Approve Advance Request"},
		{Name: "Advance Request", Entity: "advance-request", Action: "reject", ActionName: "Reject Advance Request"},
		{Name: "Advance Request", Entity: "advance-request", Action: "close", ActionName: "Manually Close Advance Request"},
		{Name: "Advance Request", Entity: "advance-request", Action: "send-to-sqlacc", ActionName: "Send To SQL Account"},
		{Name: "Advance Request", Entity: "advance-request", Action: "export", ActionName: "Export Advance Requests"},

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

		// Approver List
		{Name: "Approver List", Entity: "approver-list", Action: "view", ActionName: "View Approver List"},
		{Name: "Approver List", Entity: "approver-list", Action: "export", ActionName: "Export Approver List"},
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
