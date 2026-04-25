package dtos

type AnalyticsSpendItem struct {
	Name   string  `json:"name"`
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}

type AnalyticsLeaderboardItem struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	Team   string  `json:"team"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type AnalyticsActivityItem struct {
	Kind   string  `json:"kind"`
	Who    string  `json:"who"`
	Code   string  `json:"code"`
	Text   string  `json:"text"`
	Amount float64 `json:"amount"`
	When   string  `json:"when"`
}

type AnalyticsResponse struct {
	SpendByProject []AnalyticsSpendItem       `json:"spend_by_project"`
	SpendByGL      []AnalyticsSpendItem       `json:"spend_by_gl"`
	SpendByPayment []AnalyticsSpendItem       `json:"spend_by_payment"`
	TopSubmitters  []AnalyticsLeaderboardItem `json:"top_submitters"`
	TopApprovers   []AnalyticsLeaderboardItem `json:"top_approvers"`
	RecentActivity []AnalyticsActivityItem    `json:"recent_activity"`
}
