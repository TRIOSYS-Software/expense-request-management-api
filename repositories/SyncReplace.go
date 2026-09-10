package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SyncCounts struct {
	Upserted int64
	Deleted  int64
}

func replaceAll[T any, K comparable](
	db *gorm.DB,
	rows []T,
	keyColumn string,
	keyOf func(T) K,
) (SyncCounts, error) {
	var counts SyncCounts

	err := db.Transaction(func(tx *gorm.DB) error {
		keep := make([]K, 0, len(rows))
		for _, row := range rows {
			keep = append(keep, keyOf(row))
		}

		var model T
		del := tx.Where(keyColumn+" NOT IN ?", keep)
		if len(keep) == 0 {
			del = tx.Where("1 = 1")
		}
		delRes := del.Delete(&model)
		if delRes.Error != nil {
			return delRes.Error
		}
		counts.Deleted = delRes.RowsAffected

		if len(rows) == 0 {
			return nil
		}

		upRes := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: keyColumn}},
			UpdateAll: true,
		}).Create(&rows)
		if upRes.Error != nil {
			return upRes.Error
		}
		counts.Upserted = upRes.RowsAffected
		return nil
	})

	return counts, err
}
