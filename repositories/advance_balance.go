package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

// balanceEpsilon is a tiny float tolerance for input-validation comparisons (caps, returned bound).
const balanceEpsilon = 1e-6

// settledThreshold is the remaining balance (in Kyat) below which an advance is considered fully
// consumed. Kyat has no sub-unit in practice, so any sub-1 remainder is floating-point dust from
// MySQL's SUM/LEAST over double columns (e.g. a 1500 == 1500 settlement reporting remaining 0.0019),
// never a real balance. Anything ≥ 1 is a genuine remainder and keeps the advance open.
const settledThreshold = 1.0

const consumedSQL = `CASE
		WHEN advance_used_amount IS NOT NULL AND advance_used_amount > 0
			THEN LEAST(amount + COALESCE(returned_amount, 0), advance_used_amount)
		ELSE amount
	END`

func advanceConsumed(db *gorm.DB, advanceRequestID uint, excludeERID *uint, statuses []string) (float64, error) {
	q := db.Model(&models.ExpenseRequests{}).
		Where("advance_request_id = ? AND status IN ?", advanceRequestID, statuses)
	if excludeERID != nil {
		q = q.Where("id <> ?", *excludeERID)
	}
	var consumed *float64
	if err := q.Select("SUM(" + consumedSQL + ")").Scan(&consumed).Error; err != nil {
		return 0, err
	}
	if consumed == nil {
		return 0, nil
	}
	return *consumed, nil
}

func advanceRemaining(db *gorm.DB, ar *models.AdvanceRequests, excludeERID *uint) (float64, error) {
	consumed, err := advanceConsumed(db, ar.ID, excludeERID, []string{"pending", "approved"})
	if err != nil {
		return 0, err
	}
	remaining := ar.Amount - consumed
	if remaining < settledThreshold {
		remaining = 0
	}
	return remaining, nil
}

func fillAdvanceBalances(db *gorm.DB, ars []models.AdvanceRequests) error {
	if len(ars) == 0 {
		return nil
	}
	ids := make([]uint, len(ars))
	for i := range ars {
		ids[i] = ars[i].ID
	}

	type consumedRow struct {
		AdvanceRequestID uint
		Consumed         float64
	}
	var rows []consumedRow
	if err := db.Model(&models.ExpenseRequests{}).
		Select("advance_request_id, SUM("+consumedSQL+") AS consumed").
		Where("advance_request_id IN ? AND status IN ?", ids, []string{"pending", "approved"}).
		Group("advance_request_id").
		Scan(&rows).Error; err != nil {
		return err
	}

	consumedByID := make(map[uint]float64, len(rows))
	for _, r := range rows {
		consumedByID[r.AdvanceRequestID] = r.Consumed
	}

	for i := range ars {
		consumed := consumedByID[ars[i].ID]
		// Snap a sub-1-Kyat (dust) or over-consumed remainder to a clean, fully-settled figure.
		if ars[i].Amount-consumed < settledThreshold {
			consumed = ars[i].Amount
		}
		ars[i].SettledAmount = consumed
		ars[i].RemainingBalance = ars[i].Amount - consumed
	}
	return nil
}

func advanceFullySettled(db *gorm.DB, ar *models.AdvanceRequests) (bool, error) {
	settled, err := advanceConsumed(db, ar.ID, nil, []string{"approved"})
	if err != nil {
		return false, err
	}
	// Fully settled once less than 1 Kyat remains (sub-unit residue is float dust, not real balance).
	return ar.Amount-settled < settledThreshold, nil
}

func ReconcileAdvanceStatuses(db *gorm.DB) (int64, error) {
	res := db.Exec(`
		UPDATE advance_requests ar
		SET ar.status = 'completed'
		WHERE ar.status = 'approved'
		  AND (
			SELECT COALESCE(SUM(`+consumedSQL+`), 0)
			FROM expense_requests
			WHERE expense_requests.advance_request_id = ar.id
			  AND expense_requests.status = 'approved'
		  ) > ar.amount - 1
	`)
	return res.RowsAffected, res.Error
}
