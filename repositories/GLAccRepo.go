package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GLAccRepo struct {
	db *gorm.DB
}

func NewGLAccRepo(db *gorm.DB) *GLAccRepo {
	return &GLAccRepo{db: db}
}

var glAccRefs = []childRef{
	{Table: "expense_requests", Column: "gl_account"},
	{Table: "advance_requests", Column: "gl_account"},
	{Table: "users_gl_accounts", Column: "gl_acc_dockey"},
	{Table: "approval_policy_gl_accounts", Column: "gl_account_dockey"},
}

func (r *GLAccRepo) GetGLAcc() ([]models.GLAcc, error) {
	var glAcc []models.GLAcc
	err := r.db.Find(&glAcc).Error
	if err != nil {
		return nil, err
	}
	return glAcc, nil
}

// ReplaceGLAcc upserts the supplied set and removes any locally cached rows
// whose DOCKEY is not in the new set and is not still referenced by a request,
// an approval policy or a user assignment, all within one transaction so a
// network or DB failure can't leave a half-synced chart of accounts.
func (r *GLAccRepo) ReplaceGLAcc(glAccs []models.GLAcc) (SyncCounts, error) {
	var counts SyncCounts
	if len(glAccs) == 0 {
		return counts, nil
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		keep := make([]int, 0, len(glAccs))
		for _, a := range glAccs {
			keep = append(keep, a.DOCKEY)
		}

		deleted, retained, err := reconcile(tx, "gl_accs", "DOCKEY", keep, glAccRefs)
		if err != nil {
			return err
		}
		counts.Deleted = deleted
		counts.Retained = retained

		upRes := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "DOCKEY"}},
			UpdateAll: true,
		}).Create(&glAccs)
		if upRes.Error != nil {
			return upRes.Error
		}
		counts.Upserted = upRes.RowsAffected
		return nil
	})
	return counts, err
}
