package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type DepartmentsRepo struct {
	db *gorm.DB
}

func NewDepartmentsRepo(db *gorm.DB) *DepartmentsRepo {
	return &DepartmentsRepo{db: db}
}

func (d *DepartmentsRepo) GetDepartments() ([]models.Departments, error) {
	var departments []models.Departments
	err := d.db.Find(&departments).Error
	return departments, err
}

func (d *DepartmentsRepo) CreateDepartment(department *models.Departments) error {
	return d.db.Create(department).Error
}

func (d *DepartmentsRepo) GetDepartmentByID(id uint) (*models.Departments, error) {
	var department models.Departments
	err := d.db.First(&department, id).Error
	return &department, err
}

func (d *DepartmentsRepo) UpdateDepartment(id uint, department *models.Departments) error {
	oldDepartment, err := d.GetDepartmentByID(id)
	if err != nil {
		return err
	}
	department.ID = oldDepartment.ID
	return d.db.Save(department).Error
}

func (d *DepartmentsRepo) DeleteDepartment(id uint) error {
	return d.db.Delete(&models.Departments{}, id).Error
}
