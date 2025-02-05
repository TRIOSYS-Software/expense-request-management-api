package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type UsersRepo struct {
	db *gorm.DB
}

func NewUsersRepo(db *gorm.DB) *UsersRepo {
	return &UsersRepo{db: db}
}

func (u *UsersRepo) GetUsers() ([]models.Users, error) {
	var users []models.Users
	err := u.db.Model(&models.Users{}).Select("id, name, email, role_id, department_id, created_at, updated_at").Find(&users).Error
	// err := u.db.Find(&users).Error
	return users, err
}

func (u *UsersRepo) GetUserByID(id uint) (*models.Users, error) {
	var user models.Users
	err := u.db.Model(&models.Users{}).Select("id, name, email, role_id, department_id, created_at, updated_at").First(&user, id).Error
	return &user, err
}

func (u *UsersRepo) CreateUser(user *models.Users) error {
	return u.db.Create(user).Error
}

func (u *UsersRepo) UpdateUser(user *models.Users) error {
	return u.db.Save(user).Error
}

func (u *UsersRepo) DeleteUser(id uint) error {
	return u.db.Delete(&models.Users{}, id).Error
}
