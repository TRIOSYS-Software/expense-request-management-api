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

func (r *ProjectRepo) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepo) SaveProjects(projects []models.Project) error {
	for _, project := range projects {
		err := r.db.Save(&project).Error
		if err != nil {
			return err
		}
	}
	return nil
}
