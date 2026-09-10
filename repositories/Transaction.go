package repositories

import (
	"fmt"
	"log"

	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

func (r *ExpenseRequestsRepo) WithTx(fn func(*ExpenseRequestsRepo) error) (err error) {
	return r.db.Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("internal error in expense transaction: %v", rec)
				log.Printf("PANIC recovered in expense transaction: %v", rec)
			}
		}()

		scoped := *r
		scoped.db = tx
		if fnErr := fn(&scoped); fnErr != nil {
			return fnErr
		}
		return err
	})
}

func (r *ExpenseRequestsRepo) FindAdvanceRequest(id uint) (*models.AdvanceRequests, error) {
	var ar models.AdvanceRequests
	if err := r.db.First(&ar, id).Error; err != nil {
		return nil, err
	}
	return &ar, nil
}

func (r *ExpenseRequestsRepo) AdvanceRemainingFor(ar *models.AdvanceRequests, excludeERID *uint) (float64, error) {
	return AdvanceRemaining(r.db, ar, excludeERID)
}

func (r *ExpenseRequestsRepo) GetExpenseRequestForUpdate(id uint) (*models.ExpenseRequests, error) {
	var er models.ExpenseRequests
	if err := r.db.First(&er, id).Error; err != nil {
		return nil, err
	}
	return &er, nil
}

func (r *AdvanceRequestsRepo) WithTx(fn func(*AdvanceRequestsRepo) error) (err error) {
	return r.db.Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("internal error in advance transaction: %v", rec)
				log.Printf("PANIC recovered in advance transaction: %v", rec)
			}
		}()

		scoped := *r
		scoped.db = tx
		if fnErr := fn(&scoped); fnErr != nil {
			return fnErr
		}
		return err
	})
}

// FindRequestUser loads the submitting user with the role and department the
// approval policy lookup and notification text depend on.
func (r *ExpenseRequestsRepo) FindRequestUser(id uint) (*models.Users, error) {
	var user models.Users
	err := r.db.Preload("Roles").Preload("Departments").Where("id = ?", id).First(&user).Error
	return &user, err
}

// FindExpensePolicy resolves the approval policy governing an expense request.
func (r *ExpenseRequestsRepo) FindExpensePolicy(request *models.ExpenseRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	return findMatchingPolicy(r.db, "expense", departmentID, request.Project, request.Amount, request.GLAccount)
}

// PolicyApprovers returns a policy's approvers ordered by approval level.
func (r *ExpenseRequestsRepo) PolicyApprovers(policyID uint) ([]models.ApprovalPoliciesUsers, error) {
	var approvers []models.ApprovalPoliciesUsers
	err := r.db.Preload("Approver").Where("approval_policy_id = ?", policyID).Order("level ASC").Find(&approvers).Error
	return approvers, err
}

// InsertExpenseRequest stores a new expense request.
func (r *ExpenseRequestsRepo) InsertExpenseRequest(request *models.ExpenseRequests) error {
	return r.db.Create(request).Error
}

// InsertExpenseApproval stores one link in an approval chain.
func (r *ExpenseRequestsRepo) InsertExpenseApproval(approval *models.ExpenseApprovals) error {
	return r.db.Create(approval).Error
}

// FindAdvanceRequestUser mirrors FindRequestUser for the advance repository.
func (r *AdvanceRequestsRepo) FindRequestUser(id uint) (*models.Users, error) {
	var user models.Users
	err := r.db.Preload("Roles").Preload("Departments").Where("id = ?", id).First(&user).Error
	return &user, err
}

// FindAdvancePolicy resolves the approval policy governing an advance request.
func (r *AdvanceRequestsRepo) FindAdvancePolicy(request *models.AdvanceRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	return findMatchingPolicy(r.db, "advance", departmentID, request.Project, request.Amount, request.GLAccount)
}

// PolicyApprovers returns a policy's approvers ordered by approval level.
func (r *AdvanceRequestsRepo) PolicyApprovers(policyID uint) ([]models.ApprovalPoliciesUsers, error) {
	var approvers []models.ApprovalPoliciesUsers
	err := r.db.Preload("Approver").Where("approval_policy_id = ?", policyID).Order("level ASC").Find(&approvers).Error
	return approvers, err
}

// InsertAdvanceRequest stores a new advance request.
func (r *AdvanceRequestsRepo) InsertAdvanceRequest(request *models.AdvanceRequests) error {
	return r.db.Create(request).Error
}

// InsertAdvanceApproval stores one link in an approval chain.
func (r *AdvanceRequestsRepo) InsertAdvanceApproval(approval *models.AdvanceApprovals) error {
	return r.db.Create(approval).Error
}
