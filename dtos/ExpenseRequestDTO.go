package dtos

type DailyBreakdown struct {
	Approved  float64 `json:"approved"`
	Pending   float64 `json:"pending"`
	Rejected  float64 `json:"rejected"`
	Completed float64 `json:"completed"`
}

type ExpenseRequestSummary struct {
	Total             int                       `json:"total"`
	Pending           int                       `json:"pending"`
	Approved          int                       `json:"approved"`
	Rejected          int                       `json:"rejected"`
	Completed         int                       `json:"completed"`
	TotalAmount       float64                   `json:"total_amount"`
	ApprovedAmount    float64                   `json:"approved_amount"`
	PendingAmount     float64                   `json:"pending_amount"`
	CompletedAmount   float64                   `json:"completed_amount"`
	AdvanceUsedAmount float64                   `json:"advance_used_amount"`
	DailyTotal        map[string]DailyBreakdown `json:"daily_totals"`
}

type CompleteExpenseRequestDTO struct {
	Comment *string `json:"comment"`
}
