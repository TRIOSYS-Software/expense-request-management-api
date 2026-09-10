package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

// findMatchingPolicy resolves the approval policy that governs a request.
//
// A policy matches when all of the following hold:
//
//   - it is of the requested type ("expense" or "advance");
//   - it targets the requester's department, or is department-agnostic
//     (department_id IS NULL);
//   - its project matches exactly;
//   - the amount falls within [min_amount, max_amount];
//   - it either lists no GL accounts at all, or lists the request's GL account.
func findMatchingPolicy(
	tx *gorm.DB,
	policyType string,
	departmentID uint,
	project string,
	amount float64,
	glAccount string,
) (*models.ApprovalPolicies, error) {
	const glAccountsForPolicy = `SELECT 1 FROM approval_policy_gl_accounts
		WHERE approval_policy_id = approval_policies.id`

	var policy models.ApprovalPolicies
	err := tx.Where(
		`policy_type = ?
		 AND (department_id = ? OR department_id IS NULL)
		 AND project = ?
		 AND ? BETWEEN min_amount AND max_amount
		 AND (
			NOT EXISTS (`+glAccountsForPolicy+`)
			OR EXISTS (`+glAccountsForPolicy+` AND gl_account_dockey = CAST(? AS UNSIGNED))
		 )`,
		policyType, departmentID, project, amount, glAccount,
	).
		Order("NOT EXISTS (" + glAccountsForPolicy + ") ASC").
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}
