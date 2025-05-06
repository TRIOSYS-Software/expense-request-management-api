package services

import (
	"errors"
	"fmt"
	helper "shwetaik-expense-management-api/Helper"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/dtos"
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

func (u *UsersService) GetUsersByRole(roleID uint) (*[]models.Users, error) {
	return u.UsersRepo.GetUsersByRole(roleID)
}

func (u *UsersService) UpdateUser(user *models.Users) error {
	if user.Password != "" {
		hashPassword, err := helper.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashPassword
	}
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

func (u *UsersService) SetPaymentMethodsToUser(request *dtos.UserPaymentMethodDTO) error {
	return u.UsersRepo.SetPaymentMethodsToUser(request)
}

func (u *UsersService) GetUsersWithPaymentMethods() (*[]models.Users, error) {
	return u.UsersRepo.GetUsersWithPaymentMethods()
}

func (u *UsersService) GetUserPaymentMethods(userID uint) (*[]models.PaymentMethod, error) {
	return u.UsersRepo.GetUserPaymentMethods(userID)
}

func (u *UsersService) SetGLAccountsToUser(request *dtos.UserGLAccountDTO) error {
	return u.UsersRepo.SetGLAccountsToUser(request)
}

func (u *UsersService) GetUsersWithGLAccounts() (*[]models.Users, error) {
	return u.UsersRepo.GetUsersWithGLAccounts()
}

func (u *UsersService) GetUserGLAccounts(userID uint) (*[]models.GLAcc, error) {
	return u.UsersRepo.GetUserGLAccounts(userID)
}

func (u *UsersService) ChangePassword(id uint, request *dtos.ChangePasswordRequestDTO) error {
	user, err := u.UsersRepo.GetUserByID(id)
	if err != nil {
		return err
	}
	if !helper.CheckPasswordHash(request.OldPassword, user.Password) {
		return fmt.Errorf("invalid old password")
	}
	hashPassword, err := helper.HashPassword(request.NewPassword)
	if err != nil {
		return err
	}
	user.Password = hashPassword
	return u.UsersRepo.UpdateUser(user)
}

func (u *UsersService) ForgotPassword(request *dtos.PasswordResetRequestDTO) error {
	user, err := u.UsersRepo.GetUserByEmail(request.Email)
	if err != nil {
		return err
	}
	token := helper.GenerateToken()
	if token == "" {
		return errors.New("failed to generate token")
	}
	passwordReset := &models.PasswordReset{
		UserID:    user.ID,
		Token:     token,
		ExpiredAt: time.Now().Add(time.Minute * 5),
		Used:      false,
	}
	if err := u.UsersRepo.CreatePasswordReset(passwordReset); err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s", configs.Envs.FRONTEND_URL, token, user.Email)
	mailBody := fmt.Sprintf("Please click the link below to reset your password:\n\n%s", resetLink)

	if err := helper.SendEmail([]string{user.Email}, "Password Reset", mailBody); err != nil {
		return err
	}

	return nil
}

func (u *UsersService) ValidatePasswordResetToken(token dtos.PasswordResetTokenDTO) error {
	passwordReset := models.PasswordReset{}
	if err := u.UsersRepo.ValidatePasswordResetToken(&passwordReset, token); err != nil {
		return err
	}
	if passwordReset.ExpiredAt.Before(time.Now()) {
		return errors.New("password reset token has expired")
	}
	if passwordReset.Used {
		return errors.New("password reset token has already been used")
	}
	return nil
}

func (u *UsersService) ResetPassword(request *dtos.PasswordResetChangeRequestDTO) error {
	user, err := u.UsersRepo.GetUserByEmail(request.Email)
	if err != nil {
		return err
	}
	hashPassword, err := helper.HashPassword(request.Password)
	if err != nil {
		return err
	}
	user.Password = hashPassword
	if err = u.UsersRepo.UpdateUser(user); err != nil {
		return err
	}
	if err = u.UsersRepo.DeletePasswordReset(&models.PasswordReset{UserID: user.ID}); err != nil {
		return err
	}
	return nil
}
