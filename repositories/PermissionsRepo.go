package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type PermissionsRepo struct {
	DB *gorm.DB
}

func NewPermissionsRepo(db *gorm.DB) *PermissionsRepo {
	return &PermissionsRepo{DB: db}
}

func (p *PermissionsRepo) GetAll() ([]models.Permissions, error) {
	var permissions []models.Permissions
	if err := p.DB.Order("entity, id").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (p *PermissionsRepo) GetByIDs(ids []uint) ([]models.Permissions, error) {
	var perms []models.Permissions
	if len(ids) == 0 {
		return perms, nil
	}
	if err := p.DB.Where("id IN ?", ids).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (p *PermissionsRepo) HasPermission(roleID uint, entity, action string) (bool, error) {
	var count int64
	err := p.DB.Table("roles_permissions").
		Joins("JOIN permissions ON roles_permissions.permission_id = permissions.id").
		Where("roles_permissions.role_id = ? AND permissions.entity = ? AND permissions.action = ?", roleID, entity, action).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}