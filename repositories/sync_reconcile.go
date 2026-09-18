package repositories

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type childRef struct {
	Table  string
	Column string
}

func reconcile(tx *gorm.DB, table, keyCol string, keep any, refs []childRef) (int64, []string, error) {
	stale := fmt.Sprintf("`%s`.`%s` NOT IN (?)", table, keyCol)

	conds := make([]string, 0, len(refs)+1)
	conds = append(conds, stale)
	for _, r := range refs {
		conds = append(conds, fmt.Sprintf(
			"NOT EXISTS (SELECT 1 FROM `%s` WHERE `%s`.`%s` = `%s`.`%s`)",
			r.Table, r.Table, r.Column, table, keyCol,
		))
	}

	delRes := tx.Exec(
		fmt.Sprintf("DELETE FROM `%s` WHERE %s", table, strings.Join(conds, " AND ")),
		keep,
	)
	if delRes.Error != nil {
		return 0, nil, delRes.Error
	}

	retained := []string{}
	if err := tx.Table(table).Where(stale, keep).Pluck(keyCol, &retained).Error; err != nil {
		return delRes.RowsAffected, nil, err
	}
	return delRes.RowsAffected, retained, nil
}
