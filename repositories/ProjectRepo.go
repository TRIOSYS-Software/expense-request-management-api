package repositories

import (
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

func (r *ProjectRepo) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// SaveProjects upserts the supplied set and removes any locally cached rows
// whose CODE is not in the new set, all within one transaction.
func (r *ProjectRepo) SaveProjects(projects []models.Project) (SyncCounts, error) {
	var counts SyncCounts
	err := r.db.Transaction(func(tx *gorm.DB) error {
		keep := make([]string, 0, len(projects))
		for _, p := range projects {
			keep = append(keep, p.CODE)
		}

		del := tx.Where("CODE NOT IN ?", keep)
		if len(keep) == 0 {
			del = tx.Where("1 = 1")
		}
		delRes := del.Delete(&models.Project{})
		if delRes.Error != nil {
			return delRes.Error
		}
		counts.Deleted = delRes.RowsAffected

		if len(projects) == 0 {
			return nil
		}
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
