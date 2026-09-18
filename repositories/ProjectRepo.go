package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
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

func (r *ProjectRepo) SaveProjects(projects []models.Project) (SyncCounts, error) {
	return replaceAll(r.db, projects, "projects", "CODE", func(p models.Project) string { return p.CODE }, projectRefs)
}
