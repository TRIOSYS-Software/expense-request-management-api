package dtos

type DailyBreakdown struct {
	Approved float64 `json:"approved"`
	Pending  float64 `json:"pending"`
	Rejected float64 `json:"rejected"`
}

type ExpenseRequestSummary struct {
	Total       int                       `json:"total"`
	Pending     int                       `json:"pending"`
	Approved    int                       `json:"approved"`
	Rejected    int                       `json:"rejected"`
	TotalAmount float64                   `json:"total_amount"`
	DailyTotal  map[string]DailyBreakdown `json:"daily_totals"`
}
