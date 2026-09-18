package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SyncCounts is what a transactional reconcile-and-upsert reports back to
// the caller: how many rows were affected by each step, plus the keys that
// disappeared upstream but are still referenced locally and so were kept.
type SyncCounts struct {
	Upserted int64
	Deleted  int64
	Retained []string
}

// replaceAll upserts the supplied set and removes any locally cached rows whose
// key is not in the new set and is not still referenced by one of refs, all
// within one transaction so a network or DB failure can't leave a half-synced
// table. An empty set is treated as a failed fetch and changes nothing.
func replaceAll[T any, K comparable](
	db *gorm.DB,
	rows []T,
	table string,
	keyColumn string,
	keyOf func(T) K,
	refs []childRef,
) (SyncCounts, error) {
	var counts SyncCounts
	if len(rows) == 0 {
		return counts, nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		keep := make([]K, 0, len(rows))
		for _, row := range rows {
			keep = append(keep, keyOf(row))
		}

		deleted, retained, err := reconcile(tx, table, keyColumn, keep, refs)
		if err != nil {
			return err
		}
		counts.Deleted = deleted
		counts.Retained = retained

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
