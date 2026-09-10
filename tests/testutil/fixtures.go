package testutil

import (
	"testing"
	"time"

	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type Baseline struct {
	Role          models.Roles
	AdminRole     models.Roles
	Department    models.Departments
	User          models.Users
	Approver      models.Users
	Project       models.Project
	PaymentMethod models.PaymentMethod
	GLAcc         models.GLAcc
}

func SeedBaseline(t *testing.T, db *gorm.DB) Baseline {
	t.Helper()

	b := Baseline{
		Role:          models.Roles{Name: "Staff", Description: "Standard user"},
		AdminRole:     models.Roles{Name: "Admin", Description: "Administrator", IsAdmin: true},
		Department:    models.Departments{Name: "Engineering"},
		Project:       models.Project{CODE: "PRJ-1"},
		PaymentMethod: models.PaymentMethod{CODE: "CASH", DESCRIPTION: "Cash Account"},
		GLAcc:         models.GLAcc{DOCKEY: 5001, CODE: "6000/000", DESCRIPTION: "Office Expenses"},
	}

	mustCreate(t, db, &b.Role)
	mustCreate(t, db, &b.AdminRole)
	mustCreate(t, db, &b.Department)
	mustCreate(t, db, &b.Project)
	mustCreate(t, db, &b.PaymentMethod)
	mustCreate(t, db, &b.GLAcc)

	b.User = models.Users{
		Name:         "Requester",
		Email:        "requester@example.com",
		Password:     "hashed",
		RoleID:       b.Role.ID,
		DepartmentID: &b.Department.ID,
	}
	b.Approver = models.Users{
		Name:         "Approver",
		Email:        "approver@example.com",
		Password:     "hashed",
		RoleID:       b.Role.ID,
		DepartmentID: &b.Department.ID,
	}
	mustCreate(t, db, &b.User)
	mustCreate(t, db, &b.Approver)

	return b
}

func mustCreate(t *testing.T, db *gorm.DB, value any) {
	t.Helper()
	if err := db.Create(value).Error; err != nil {
		t.Fatalf("seed %T: %v", value, err)
	}
}

// NewUser inserts an extra user against the baseline's role and department.
func (b Baseline) NewUser(t *testing.T, db *gorm.DB, name, email string) models.Users {
	t.Helper()
	u := models.Users{
		Name:         name,
		Email:        email,
		Password:     "hashed",
		RoleID:       b.Role.ID,
		DepartmentID: &b.Department.ID,
	}
	mustCreate(t, db, &u)
	return u
}

// NewAdvance inserts an advance request owned by the baseline user.
func (b Baseline) NewAdvance(t *testing.T, db *gorm.DB, amount float64, status string) models.AdvanceRequests {
	t.Helper()
	ar := models.AdvanceRequests{
		Amount:               amount,
		Description:          "advance",
		Project:              b.Project.CODE,
		PaymentMethod:        b.PaymentMethod.CODE,
		UserID:               b.User.ID,
		GLAccount:            "5001",
		DateSubmitted:        time.Now(),
		Status:               status,
		CurrentApproverLevel: 1,
	}
	mustCreate(t, db, &ar)
	return ar
}

type ExpenseOpts struct {
	Amount           float64
	Status           string
	AdvanceRequestID *uint
	AdvanceUsed      *float64
	Returned         *float64
	UserID           *uint
	DateSubmitted    *time.Time
}

func (b Baseline) NewExpense(t *testing.T, db *gorm.DB, opts ExpenseOpts) models.ExpenseRequests {
	t.Helper()

	userID := b.User.ID
	if opts.UserID != nil {
		userID = *opts.UserID
	}
	submitted := time.Now()
	if opts.DateSubmitted != nil {
		submitted = *opts.DateSubmitted
	}
	status := opts.Status
	if status == "" {
		status = "pending"
	}

	er := models.ExpenseRequests{
		Amount:               opts.Amount,
		Description:          "expense",
		Project:              b.Project.CODE,
		PaymentMethod:        b.PaymentMethod.CODE,
		UserID:               userID,
		GLAccount:            "5001",
		DateSubmitted:        submitted,
		Status:               status,
		CurrentApproverLevel: 1,
		AdvanceRequestID:     opts.AdvanceRequestID,
		AdvanceUsedAmount:    opts.AdvanceUsed,
		ReturnedAmount:       opts.Returned,
	}
	mustCreate(t, db, &er)
	return er
}

// Ptr returns a pointer to v, for the optional fields above.
func Ptr[T any](v T) *T { return &v }


func (b Baseline) NewApprovalChain(t *testing.T, db *gorm.DB, requestID uint, approverIDs ...uint) []models.ExpenseApprovals {
	t.Helper()
	out := make([]models.ExpenseApprovals, 0, len(approverIDs))
	for i, approverID := range approverIDs {
		a := models.ExpenseApprovals{
			RequestID:  requestID,
			ApproverID: approverID,
			Level:      uint(i + 1),
			Status:     "pending",
			IsFinal:    i == len(approverIDs)-1,
		}
		mustCreate(t, db, &a)
		out = append(out, a)
	}
	return out
}

func (b Baseline) NewAdvanceApprovalChain(t *testing.T, db *gorm.DB, requestID uint, approverIDs ...uint) []models.AdvanceApprovals {
	t.Helper()
	out := make([]models.AdvanceApprovals, 0, len(approverIDs))
	for i, approverID := range approverIDs {
		a := models.AdvanceApprovals{
			RequestID:  requestID,
			ApproverID: approverID,
			Level:      uint(i + 1),
			Status:     "pending",
			IsFinal:    i == len(approverIDs)-1,
		}
		mustCreate(t, db, &a)
		out = append(out, a)
	}
	return out
}
