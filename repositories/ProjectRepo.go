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

var projectRefs = []childRef{
	{Table: "approval_policies", Column: "project"},
	{Table: "expense_requests", Column: "project"},
	{Table: "advance_requests", Column: "project"},
	{Table: "users_projects", Column: "project_code"},
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
// whose CODE is not in the new set and is not still referenced by a request,
// an approval policy or a user assignment, all within one transaction.
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

		deleted, retained, err := reconcile(tx, "projects", "CODE", keep, projectRefs)
		if err != nil {
			return err
		}
		counts.Deleted = deleted
		counts.Retained = retained

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
