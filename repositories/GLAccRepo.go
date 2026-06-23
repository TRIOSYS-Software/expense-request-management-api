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

func (r *GLAccRepo) GetGLAcc() ([]models.GLAcc, error) {
	var glAcc []models.GLAcc
	err := r.db.Find(&glAcc).Error
	if err != nil {
		return nil, err
	}
	return glAcc, nil
}

// ReplaceGLAcc upserts the supplied set and removes any locally cached rows
// whose DOCKEY is not in the new set, all within one transaction so a
// network or DB failure can't leave a half-synced chart of accounts.
func (r *GLAccRepo) ReplaceGLAcc(glAccs []models.GLAcc) (SyncCounts, error) {
	var counts SyncCounts
	err := r.db.Transaction(func(tx *gorm.DB) error {
		keep := make([]int, 0, len(glAccs))
		for _, a := range glAccs {
			keep = append(keep, a.DOCKEY)
		}

		del := tx.Where("DOCKEY NOT IN ?", keep)
		if len(keep) == 0 {
			del = tx.Where("1 = 1")
		}
		delRes := del.Delete(&models.GLAcc{})
		if delRes.Error != nil {
			return delRes.Error
		}
		counts.Deleted = delRes.RowsAffected

		if len(glAccs) == 0 {
			return nil
		}
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
