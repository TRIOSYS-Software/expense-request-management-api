package services

import (
	"errors"
	helper "shwetaik-expense-management-api/Helper"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"

	"time"

	"github.com/golang-jwt/jwt/v5"
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

func (u *UsersService) LoginUser(user *models.Users) (*struct {
	User  models.Users
	Token string
}, error) {
	getUser, err := u.UsersRepo.LoginUser(user)
	if err != nil {
		return nil, err
	}
	if !helper.CheckPasswordHash(user.Password, getUser.Password) {
		return nil, errors.New("invalid password")
	}
	claims := jwt.MapClaims{
		"user_id":    getUser.ID,
		"user_name":  getUser.Name,
		"user_email": getUser.Email,
		"user_role":  getUser.RoleID,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(configs.Envs.JWTSecret))
	if err != nil {
		return nil, err
	}

	data := &struct {
		User  models.Users
		Token string
	}{
		User:  *getUser,
		Token: tokenString,
	}

	return data, err
}
