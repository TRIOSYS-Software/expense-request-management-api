package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type DepartmentsService struct {
	DepartmentsRepo *repositories.DepartmentsRepo
}

func NewDepartmentsService(departmentsRepo *repositories.DepartmentsRepo) *DepartmentsService {
	return &DepartmentsService{DepartmentsRepo: departmentsRepo}
}

func (d *DepartmentsService) GetDepartments() ([]models.Departments, error) {
	return d.DepartmentsRepo.GetDepartments()
}

func (d *DepartmentsService) CreateDepartment(department *models.Departments) error {
	return d.DepartmentsRepo.CreateDepartment(department)
}

func (d *DepartmentsService) GetDepartmentByID(id uint) (*models.Departments, error) {
	return d.DepartmentsRepo.GetDepartmentByID(id)
}

func (d *DepartmentsService) UpdateDepartment(id uint, department *models.Departments) error {
	return d.DepartmentsRepo.UpdateDepartment(id, department)
}

func (d *DepartmentsService) DeleteDepartment(id uint) error {
	return d.DepartmentsRepo.DeleteDepartment(id)
}
