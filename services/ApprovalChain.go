package services

import (
	"fmt"
	"log"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"

	"gorm.io/gorm"
)
const unknownRoleName = "Unknown Role"

type requestSubmitter struct {
	user         models.Users
	roleName     string
	departmentID uint
}

func resolveSubmitter(user *models.Users, loadErr error, userID uint) (*requestSubmitter, error) {
	if loadErr != nil {
		if loadErr == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("User with ID %d not found", userID)
		}
		return nil, fmt.Errorf("Failed to retrieve user: %w", loadErr)
	}

	if user.DepartmentID == nil {
		return nil, fmt.Errorf("User (ID %d - %s) has no department assigned", user.ID, user.Name)
	}

	roleName := unknownRoleName
	if user.Roles != nil {
		roleName = user.Roles.Name
	} else {
		log.Printf("WARN: User %d (%s) has no role assigned or role not found for role_id: %d", user.ID, user.Name, user.RoleID)
	}

	return &requestSubmitter{user: *user, roleName: roleName, departmentID: *user.DepartmentID}, nil
}

func createExpenseWithChain(repo *repositories.ExpenseRequestsRepo, request *models.ExpenseRequests) ([]repositories.PendingNotification, error) {
	user, err := repo.FindRequestUser(request.UserID)
	submitter, err := resolveSubmitter(user, err, request.UserID)
	if err != nil {
		return nil, err
	}

	policy, err := repo.FindExpensePolicy(request, submitter.departmentID)
	if err != nil {
		return nil, fmt.Errorf("No approval policy found")
	}

	approvers, err := repo.PolicyApprovers(policy.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve approver users: %w", err)
	}
	if len(approvers) == 0 {
		return nil, fmt.Errorf("No approver users found")
	}

	if err := repo.InsertExpenseRequest(request); err != nil {
		return nil, err
	}

	var pending []repositories.PendingNotification
	for i, approver := range approvers {
		approval := models.ExpenseApprovals{
			RequestID:  request.ID,
			ApproverID: approver.UserID,
			Level:      approver.Level,
			Status:     "pending",
			IsFinal:    i == len(approvers)-1,
		}
		if err := repo.InsertExpenseApproval(&approval); err != nil {
			return nil, err
		}

		if approver.Level == request.CurrentApproverLevel {
			pending = append(pending, repositories.PendingNotification{
				UserID: approver.UserID,
				Message: fmt.Sprintf(
					"%s (%s) has created a new expense request (#%d) for your approval. Amount: $%.2f",
					submitter.user.Name, submitter.roleName, request.ID, request.Amount,
				),
				Type: "new_request",
			})
		}
	}

	if len(pending) == 0 {
		log.Printf("WARN: No approver matched CurrentApproverLevel %d for expense request %d", request.CurrentApproverLevel, request.ID)
	}

	return pending, nil
}

func createAdvanceWithChain(repo *repositories.AdvanceRequestsRepo, request *models.AdvanceRequests) ([]repositories.PendingNotification, error) {
	user, err := repo.FindRequestUser(request.UserID)
	submitter, err := resolveSubmitter(user, err, request.UserID)
	if err != nil {
		return nil, err
	}

	policy, err := repo.FindAdvancePolicy(request, submitter.departmentID)
	if err != nil {
		return nil, fmt.Errorf("No advance approval policy found")
	}

	approvers, err := repo.PolicyApprovers(policy.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve approver users: %w", err)
	}
	if len(approvers) == 0 {
		return nil, fmt.Errorf("No approver users found")
	}

	if err := repo.InsertAdvanceRequest(request); err != nil {
		return nil, err
	}

	var pending []repositories.PendingNotification
	for i, approver := range approvers {
		approval := models.AdvanceApprovals{
			RequestID:  request.ID,
			ApproverID: approver.UserID,
			Level:      approver.Level,
			Status:     "pending",
			IsFinal:    i == len(approvers)-1,
		}
		if err := repo.InsertAdvanceApproval(&approval); err != nil {
			return nil, err
		}

		if approver.Level == request.CurrentApproverLevel {
			pending = append(pending, repositories.PendingNotification{
				UserID: approver.UserID,
				Message: fmt.Sprintf(
					"%s (%s) has created a new advance request (#%d) for your approval. Amount: $%.2f",
					submitter.user.Name, submitter.roleName, request.ID, request.Amount,
				),
				Type: "advance_new_request",
			})
		}
	}

	return pending, nil
}
