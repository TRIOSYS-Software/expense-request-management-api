package repositories

import (
	"fmt"
	"time"

	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

func (r *AdvanceRequestsRepo) advanceSummaryScope(filters map[string]any) *gorm.DB {
	db := r.db.Model(&models.AdvanceRequests{})

	if filters["need_my_approval"] != nil && filters["approver_id"] != nil {
		db = db.Joins("JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
			Where("advance_approvals.approver_id = ?", filters["approver_id"]).
			Where("advance_requests.status = 'pending'").
			Where("advance_approvals.level = advance_requests.current_approver_level").
			Group("advance_requests.id")
	} else if filters["user_id"] != nil && filters["approver_id"] != nil {
		db = db.Joins("LEFT JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
			Where("(advance_requests.user_id = ? OR advance_approvals.approver_id = ?)", filters["user_id"], filters["approver_id"]).
			Group("advance_requests.id")
	} else if filters["user_id"] != nil {
		db = db.Where("advance_requests.user_id = ?", filters["user_id"])
	} else if filters["approver_id"] != nil {
		db = db.Joins("JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
			Where("advance_approvals.approver_id = ?", filters["approver_id"]).
			Group("advance_requests.id")
	}

	if filters["status"] != nil {
		db = db.Where("advance_requests.status = ?", filters["status"].(string))
	}

	return applySummaryFilters(db, "advance_requests", filters)
}

func (r *AdvanceRequestsRepo) advanceSummaryRows(filters map[string]any) *gorm.DB {
	return r.advanceSummaryScope(filters).Select(
		"advance_requests.id AS id",
		"advance_requests.status AS status",
		"advance_requests.amount AS amount",
		"advance_requests.date_submitted AS date_submitted",
	)
}

func (r *AdvanceRequestsRepo) consumedByAdvance() *gorm.DB {
	return r.db.Model(&models.ExpenseRequests{}).
		Select("advance_request_id AS advance_request_id, SUM("+consumedSQL+") AS consumed").
		Where("status IN ?", []string{"pending", "approved"}).
		Group("advance_request_id")
}

func (r *AdvanceRequestsRepo) settledByAdvance() *gorm.DB {
	return r.db.Model(&models.ExpenseRequests{}).
		Select("advance_request_id AS advance_request_id, SUM(advance_used_amount) AS settled").
		Where("status = ?", "approved").
		Group("advance_request_id")
}

var remainingExpr = fmt.Sprintf(
	"CASE WHEN t.amount - COALESCE(c.consumed, 0) < %v THEN 0 ELSE t.amount - COALESCE(c.consumed, 0) END",
	settledThreshold,
)

func (r *AdvanceRequestsRepo) GetAdvanceRequestsSummary(filters map[string]any) (dtos.AdvanceRequestSummary, error) {
	var summary dtos.AdvanceRequestSummary

	var totals struct {
		Total           int64
		TotalAmount     float64
		Pending         int64
		Approved        int64
		Rejected        int64
		Completed       int64
		PendingAmount   float64
		ApprovedAmount  float64
		CompletedAmount float64
		RemainingAmount float64
		SettledAmount   float64
	}

	if err := r.db.Table("(?) AS t", r.advanceSummaryRows(filters)).
		Joins("LEFT JOIN (?) AS c ON c.advance_request_id = t.id", r.consumedByAdvance()).
		Joins("LEFT JOIN (?) AS s ON s.advance_request_id = t.id", r.settledByAdvance()).
		Select(`
			COUNT(*)                                                              AS total,
			COALESCE(SUM(t.amount), 0)                                            AS total_amount,
			COALESCE(SUM(t.status = 'pending'), 0)                                AS pending,
			COALESCE(SUM(t.status = 'approved'), 0)                               AS approved,
			COALESCE(SUM(t.status = 'rejected'), 0)                               AS rejected,
			COALESCE(SUM(t.status = 'completed'), 0)                              AS completed,
			COALESCE(SUM(CASE WHEN t.status = 'pending'   THEN t.amount END), 0)  AS pending_amount,
			COALESCE(SUM(CASE WHEN t.status = 'approved'  THEN t.amount END), 0)  AS approved_amount,
			COALESCE(SUM(CASE WHEN t.status = 'completed' THEN t.amount END), 0)  AS completed_amount,
			COALESCE(SUM(CASE WHEN t.status = 'approved'  THEN ` + remainingExpr + ` END), 0) AS remaining_amount,
			COALESCE(SUM(s.settled), 0)                                           AS settled_amount`).
		Scan(&totals).Error; err != nil {
		return summary, err
	}

	summary.Total = int(totals.Total)
	summary.TotalAmount = totals.TotalAmount
	summary.Pending = int(totals.Pending)
	summary.Approved = int(totals.Approved)
	summary.Rejected = int(totals.Rejected)
	summary.Completed = int(totals.Completed)
	summary.PendingAmount = totals.PendingAmount
	summary.ApprovedAmount = totals.ApprovedAmount
	summary.CompletedAmount = totals.CompletedAmount
	summary.RemainingAmount = totals.RemainingAmount
	summary.SettledAmount = totals.SettledAmount

	if filters["start_date"] == nil || filters["end_date"] == nil {
		return summary, nil
	}

	summary.DailyTotal = make(map[string]dtos.DailyBreakdown)

	var daily []struct {
		Day      time.Time
		Approved float64
		Pending  float64
		Rejected float64
	}
	
	if err := r.db.Table("(?) AS t", r.advanceSummaryRows(filters)).
		Select(`
			DATE(t.date_submitted)                                                          AS day,
			COALESCE(SUM(CASE WHEN t.status IN ('approved','completed') THEN t.amount END), 0) AS approved,
			COALESCE(SUM(CASE WHEN t.status = 'pending'  THEN t.amount END), 0)             AS pending,
			COALESCE(SUM(CASE WHEN t.status = 'rejected' THEN t.amount END), 0)             AS rejected`).
		Group("DATE(t.date_submitted)").
		Scan(&daily).Error; err != nil {
		return summary, err
	}

	for _, row := range daily {
		summary.DailyTotal[row.Day.Format("2006-01-02")] = dtos.DailyBreakdown{
			Approved: row.Approved,
			Pending:  row.Pending,
			Rejected: row.Rejected,
		}
	}

	return summary, nil
}
