package repositories

import (
	"time"

	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

func (r *ExpenseRequestsRepo) expenseSummaryScope(filters map[string]any) *gorm.DB {
	db := r.db.Model(&models.ExpenseRequests{})

	if filters["need_my_approval"] != nil && filters["approver_id"] != nil {
		db = db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("expense_approvals.approver_id = ?", filters["approver_id"]).
			Where("expense_requests.status = 'pending'").
			Where("expense_approvals.level = expense_requests.current_approver_level").
			Group("expense_requests.id")
	} else if filters["user_id"] != nil && filters["approver_id"] != nil {
		db = db.Joins("LEFT JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("(expense_requests.user_id = ? OR expense_approvals.approver_id = ?)", filters["user_id"], filters["approver_id"]).
			Group("expense_requests.id")
	} else if filters["user_id"] != nil {
		db = db.Where("expense_requests.user_id = ?", filters["user_id"])
	} else if filters["approver_id"] != nil {
		db = db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("expense_approvals.approver_id = ?", filters["approver_id"]).
			Group("expense_requests.id")
	}

	if filters["status"] != nil {
		db = db.Where("expense_requests.status = ?", filters["status"].(string))
	}
	if filters["amount"] != nil {
		db = db.Where("amount = ?", filters["amount"])
	}

	return applySummaryFilters(db, "expense_requests", filters)
}

func (r *ExpenseRequestsRepo) expenseSummaryRows(filters map[string]any) *gorm.DB {
	return r.expenseSummaryScope(filters).Select(
		"expense_requests.id AS id",
		"expense_requests.status AS status",
		"expense_requests.amount AS amount",
		"expense_requests.advance_used_amount AS advance_used_amount",
		"expense_requests.date_submitted AS date_submitted",
	)
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsSummary(filters map[string]any) (dtos.ExpenseRequestSummary, error) {
	var summary dtos.ExpenseRequestSummary

	var totals struct {
		Total             int64
		TotalAmount       float64
		Pending           int64
		Approved          int64
		Completed         int64
		Rejected          int64
		PendingAmount     float64
		ApprovedAmount    float64
		CompletedAmount   float64
		AdvanceUsedAmount float64
	}

	if err := r.db.Table("(?) AS t", r.expenseSummaryRows(filters)).
		Select(`
			COUNT(*)                                                              AS total,
			COALESCE(SUM(t.amount), 0)                                            AS total_amount,
			COALESCE(SUM(t.status = 'pending'), 0)                                AS pending,
			COALESCE(SUM(t.status = 'approved'), 0)                               AS approved,
			COALESCE(SUM(t.status = 'completed'), 0)                              AS completed,
			COALESCE(SUM(t.status = 'rejected'), 0)                               AS rejected,
			COALESCE(SUM(CASE WHEN t.status = 'pending'   THEN t.amount END), 0)  AS pending_amount,
			COALESCE(SUM(CASE WHEN t.status = 'approved'  THEN t.amount END), 0)  AS approved_amount,
			COALESCE(SUM(CASE WHEN t.status = 'completed' THEN t.amount END), 0)  AS completed_amount,
			COALESCE(SUM(CASE WHEN t.status <> 'rejected' THEN t.advance_used_amount END), 0) AS advance_used_amount`).
		Scan(&totals).Error; err != nil {
		return summary, err
	}

	summary.Total = int(totals.Total)
	summary.TotalAmount = totals.TotalAmount
	summary.Pending = int(totals.Pending)
	summary.Approved = int(totals.Approved)
	summary.Completed = int(totals.Completed)
	summary.Rejected = int(totals.Rejected)
	summary.PendingAmount = totals.PendingAmount
	summary.ApprovedAmount = totals.ApprovedAmount
	summary.CompletedAmount = totals.CompletedAmount
	summary.AdvanceUsedAmount = totals.AdvanceUsedAmount

	if filters["start_date"] == nil || filters["end_date"] == nil {
		return summary, nil
	}

	summary.DailyTotal = make(map[string]dtos.DailyBreakdown)

	var daily []struct {
		Day       time.Time
		Approved  float64
		Pending   float64
		Rejected  float64
		Completed float64
	}
	if err := r.db.Table("(?) AS t", r.expenseSummaryRows(filters)).
		Select(`
			DATE(t.date_submitted)                                                AS day,
			COALESCE(SUM(CASE WHEN t.status = 'approved'  THEN t.amount END), 0)  AS approved,
			COALESCE(SUM(CASE WHEN t.status = 'pending'   THEN t.amount END), 0)  AS pending,
			COALESCE(SUM(CASE WHEN t.status = 'rejected'  THEN t.amount END), 0)  AS rejected,
			COALESCE(SUM(CASE WHEN t.status = 'completed' THEN t.amount END), 0)  AS completed`).
		Group("DATE(t.date_submitted)").
		Scan(&daily).Error; err != nil {
		return summary, err
	}

	for _, row := range daily {
		summary.DailyTotal[row.Day.Format("2006-01-02")] = dtos.DailyBreakdown{
			Approved:  row.Approved,
			Pending:   row.Pending,
			Rejected:  row.Rejected,
			Completed: row.Completed,
		}
	}

	return summary, nil
}
