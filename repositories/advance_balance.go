package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

const balanceEpsilon = 1e-6

const consumedSQL = `CASE
		WHEN advance_used_amount IS NOT NULL
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
	if remaining < 0 {
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
		if consumed > ars[i].Amount {
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
	return settled >= ar.Amount-balanceEpsilon, nil
}
