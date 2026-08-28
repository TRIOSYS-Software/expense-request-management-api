package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type ApproversRepo struct {
	db *gorm.DB
}

func NewApproversRepo(db *gorm.DB) *ApproversRepo {
	return &ApproversRepo{db: db}
}

type ApproverExpenseAction struct {
	models.ExpenseApprovals
	Request *models.ExpenseRequests `json:"request,omitempty"`
}

type ApproverAdvanceAction struct {
	models.AdvanceApprovals
	Request *models.AdvanceRequests `json:"request,omitempty"`
}

type ApproverActionsResult struct {
	ExpenseActions []ApproverExpenseAction `json:"expense_actions"`
	AdvanceActions []ApproverAdvanceAction `json:"advance_actions"`
}

func (r *ApproversRepo) approverEligibilitySubquery() *gorm.DB {
	return r.db.
		Table("users AS u").
		Select("DISTINCT u.id").
		Joins("LEFT JOIN roles r ON r.id = u.role_id").
		Joins("LEFT JOIN roles_permissions rp ON rp.roles_id = r.id").
		Joins("LEFT JOIN permissions p ON p.id = rp.permissions_id").
		Where("u.deleted_at IS NULL").
		Where(
			r.db.
				Where("r.is_admin = ?", true).
				Or(
					r.db.
						Where("p.entity IN ?", []string{"expense-request", "advance-request"}).
						Where("p.action IN ?", []string{"approve", "reject"}),
				),
		)
}

func (r *ApproversRepo) GetApprovers() ([]models.Users, error) {
	var users []models.Users
	err := r.db.
		Preload("Roles").
		Preload("Departments").
		Where("id IN (?)", r.approverEligibilitySubquery()).
		Order("id ASC").
		Find(&users).Error
	return users, err
}

func (r *ApproversRepo) GetApproverByID(id uint) (*models.Users, error) {
	var user models.Users
	err := r.db.
		Preload("Roles").
		Preload("Departments").
		Where("id = ? AND id IN (?)", id, r.approverEligibilitySubquery()).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ApproversRepo) IsRoleAdmin(roleID uint) bool {
	var role models.Roles
	if err := r.db.Select("id, is_admin").First(&role, roleID).Error; err != nil {
		return false
	}
	return role.IsAdmin
}

func (r *ApproversRepo) GetApproverActionsForViewer(approverID, viewerID uint, viewerIsAdmin bool) (ApproverActionsResult, error) {
	result := ApproverActionsResult{
		ExpenseActions: []ApproverExpenseAction{},
		AdvanceActions: []ApproverAdvanceAction{},
	}

	// --- Expense ---
	var expenseRows []models.ExpenseApprovals
	expenseQuery := r.db.
		Where("approver_id = ? AND status IN ?", approverID, []string{"approved", "rejected"})
	if !viewerIsAdmin {
		expenseQuery = expenseQuery.Where(
			"request_id IN (?)",
			r.db.Table("expense_approvals").
				Select("request_id").
				Where("approver_id = ?", viewerID),
		)
	}
	if err := expenseQuery.
		Order("approval_date DESC").
		Find(&expenseRows).Error; err != nil {
		return result, err
	}

	if len(expenseRows) > 0 {
		ids := make([]uint, 0, len(expenseRows))
		for _, a := range expenseRows {
			ids = append(ids, a.RequestID)
		}
		var requests []models.ExpenseRequests
		if err := r.db.
			Preload("User", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, name, email")
			}).
			Preload("Projects").
			Preload("PaymentMethods").
			Preload("GLAccounts").
			Where("id IN ?", ids).
			Find(&requests).Error; err != nil {
			return result, err
		}
		byID := make(map[uint]models.ExpenseRequests, len(requests))
		for _, req := range requests {
			byID[req.ID] = req
		}
		for _, a := range expenseRows {
			req, ok := byID[a.RequestID]
			if !ok {
				continue
			}
			copyReq := req
			row := ApproverExpenseAction{ExpenseApprovals: a, Request: &copyReq}
			result.ExpenseActions = append(result.ExpenseActions, row)
		}
	}

	// --- Advance ---
	var advanceRows []models.AdvanceApprovals
	advanceQuery := r.db.
		Where("approver_id = ? AND status IN ?", approverID, []string{"approved", "rejected"})
	if !viewerIsAdmin {
		advanceQuery = advanceQuery.Where(
			"request_id IN (?)",
			r.db.Table("advance_approvals").
				Select("request_id").
				Where("approver_id = ?", viewerID),
		)
	}
	if err := advanceQuery.
		Order("approval_date DESC").
		Find(&advanceRows).Error; err != nil {
		return result, err
	}

	if len(advanceRows) > 0 {
		ids := make([]uint, 0, len(advanceRows))
		for _, a := range advanceRows {
			ids = append(ids, a.RequestID)
		}
		var requests []models.AdvanceRequests
		if err := r.db.
			Preload("User", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, name, email")
			}).
			Preload("Projects").
			Preload("PaymentMethods").
			Preload("GLAccounts").
			Where("id IN ?", ids).
			Find(&requests).Error; err != nil {
			return result, err
		}
		byID := make(map[uint]models.AdvanceRequests, len(requests))
		for _, req := range requests {
			byID[req.ID] = req
		}
		for _, a := range advanceRows {
			req, ok := byID[a.RequestID]
			if !ok {
				continue
			}
			copyReq := req
			row := ApproverAdvanceAction{AdvanceApprovals: a, Request: &copyReq}
			result.AdvanceActions = append(result.AdvanceActions, row)
		}
	}

	return result, nil
}
