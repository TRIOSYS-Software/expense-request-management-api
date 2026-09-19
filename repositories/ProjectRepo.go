package repositories

import (
	"fmt"
	"regexp"

	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

var projectRefColumns = []struct {
	Table  string
	Column string
}{
	{"approval_policies", "project"},
	{"expense_requests", "project"},
	{"advance_requests", "project"},
	{"users_projects", "project_code"},
}

var projectPrefixRe = regexp.MustCompile(`^\d+(?:-\d+)*`)

func (r *ProjectRepo) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func detectRenames(stale, keep []string) (map[string]string, []string) {
	byPrefix := make(map[string][]string)
	for _, code := range keep {
		if p := projectPrefixRe.FindString(code); p != "" {
			byPrefix[p] = append(byPrefix[p], code)
		}
	}

	renames := make(map[string]string)
	var ambiguous []string
	for _, old := range stale {
		p := projectPrefixRe.FindString(old)
		if p == "" {
			continue
		}
		candidates := byPrefix[p]
		switch {
		case len(candidates) == 1 && candidates[0] != old:
			renames[old] = candidates[0]
		case len(candidates) > 1:
			ambiguous = append(ambiguous, old)
		}
	}
	return renames, ambiguous
}

func repointProjectRefs(tx *gorm.DB, renames map[string]string) error {
	for oldCode, newCode := range renames {
		for _, ref := range projectRefColumns {
			if ref.Table == "users_projects" {
				if err := tx.Exec(
					"UPDATE IGNORE users_projects SET project_code = ? WHERE project_code = ?",
					newCode, oldCode).Error; err != nil {
					return err
				}
				if err := tx.Exec(
					"DELETE FROM users_projects WHERE project_code = ?",
					oldCode).Error; err != nil {
					return err
				}
				continue
			}
			if err := tx.Exec(
				fmt.Sprintf("UPDATE `%s` SET `%s` = ? WHERE `%s` = ?", ref.Table, ref.Column, ref.Column),
				newCode, oldCode).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ProjectRepo) SaveProjects(projects []models.Project) (SyncCounts, error) {
	var counts SyncCounts
	if len(projects) == 0 {
		return counts, nil
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		keep := make([]string, 0, len(projects))
		for _, p := range projects {
			keep = append(keep, p.CODE)
		}

		var stale []string
		if err := tx.Table("projects").Where("CODE NOT IN ?", keep).Pluck("CODE", &stale).Error; err != nil {
			return err
		}

		renames, ambiguous := detectRenames(stale, keep)
		if err := repointProjectRefs(tx, renames); err != nil {
			return err
		}
		counts.Renamed = renames
		counts.RenameSkipped = ambiguous

		delRes := tx.Where("CODE NOT IN ?", keep).Delete(&models.Project{})
		if delRes.Error != nil {
			return delRes.Error
		}
		counts.Deleted = delRes.RowsAffected

		upRes := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "CODE"}},
			UpdateAll: true,
		}).Create(&projects)
		if upRes.Error != nil {
			return upRes.Error
		}
		counts.Upserted = upRes.RowsAffected
		return nil
	})
	return counts, err
}
