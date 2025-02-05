package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type RolesRepo struct {
	db *gorm.DB
}

func NewRolesRepo(db *gorm.DB) *RolesRepo {
	return &RolesRepo{db: db}
}

func (r *RolesRepo) GetRoles() []models.Roles {
	var Roles []models.Roles
	r.db.Find(&Roles)
	return Roles
}

func (r *RolesRepo) CreateRole(role *models.Roles) error {
	return r.db.Create(role).Error
}

func (r *RolesRepo) GetRoleByID(id uint) (*models.Roles, error) {
	var role models.Roles
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *RolesRepo) GetRoleByName(name string) (*models.Roles, error) {
	var role models.Roles
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&role).Error
	return &role, err
}

func (r *RolesRepo) UpdateRole(role *models.Roles) error {
	return r.db.Save(role).Error
}

func (r *RolesRepo) DeleteRole(id uint) error {
	return r.db.Delete(&models.Roles{}, id).Error
}
