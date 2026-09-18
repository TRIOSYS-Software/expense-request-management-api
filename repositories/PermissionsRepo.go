package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type PermissionsRepo struct {
	db *gorm.DB
}

func NewPermissionsRepo(db *gorm.DB) *PermissionsRepo {
	return &PermissionsRepo{db: db}
}

func (p *PermissionsRepo) GetAll() ([]models.Permissions, error) {
	var permissions []models.Permissions
	if err := p.db.Order("entity, id").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (p *PermissionsRepo) GetByIDs(ids []uint) ([]models.Permissions, error) {
	var perms []models.Permissions
	if len(ids) == 0 {
		return perms, nil
	}
	if err := p.db.Where("id IN ?", ids).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}
