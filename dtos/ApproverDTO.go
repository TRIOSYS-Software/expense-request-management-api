package dtos

import "shwetaik-expense-management-api/models"

// ApproverExpenseAction is one expense approval an approver has acted on,
// paired with the request it belongs to.
type ApproverExpenseAction struct {
	models.ExpenseApprovals
	Request *models.ExpenseRequests `json:"request,omitempty"`
}

// ApproverAdvanceAction is the advance-request counterpart.
type ApproverAdvanceAction struct {
	models.AdvanceApprovals
	Request *models.AdvanceRequests `json:"request,omitempty"`
}

// ApproverActionsResult is the response of GET /approvers/{id}/actions.
type ApproverActionsResult struct {
	ExpenseActions []ApproverExpenseAction `json:"expense_actions"`
	AdvanceActions []ApproverAdvanceAction `json:"advance_actions"`
}
