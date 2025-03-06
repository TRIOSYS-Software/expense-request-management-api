package dtos

type ApprovedExpenseRequestsDTO struct {
	ExpenseID     uint   `json:"expense_id"`
	PaymentMethod string `json:"payment_method"`
}
