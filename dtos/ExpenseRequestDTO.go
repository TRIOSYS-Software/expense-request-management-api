package dtos

type ExpenseRequestSummary struct {
	Total       int                `json:"total"`
	Pending     int                `json:"pending"`
	Approved    int                `json:"approved"`
	Rejected    int                `json:"rejected"`
	TotalAmount float64            `json:"total_amount"`
	DailyTotal  map[string]float64 `json:"daily_totals"`
}
