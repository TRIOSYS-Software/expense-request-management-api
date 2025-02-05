package services

import (
	helper "shwetaik-expense-management-api/Helper"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type UsersService struct {
	UsersRepo *repositories.UsersRepo
}

func NewUsersService(usersRepo *repositories.UsersRepo) *UsersService {
	return &UsersService{UsersRepo: usersRepo}
}

func (u *UsersService) GetUsers() ([]models.Users, error) {
	return u.UsersRepo.GetUsers()
}

func (u *UsersService) CreateUser(user *models.Users) error {
	hashPassword, err := helper.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashPassword
	return u.UsersRepo.CreateUser(user)
}

func (u *UsersService) GetUserByID(id uint) (*models.Users, error) {
	return u.UsersRepo.GetUserByID(id)
}

func (u *UsersService) UpdateUser(user *models.Users) error {
	return u.UsersRepo.UpdateUser(user)
}

func (u *UsersService) DeleteUser(id uint) error {
	return u.UsersRepo.DeleteUser(id)
}
