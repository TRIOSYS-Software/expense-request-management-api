package repositories

import (
	"fmt"
	"shwetaik-expense-management-api/dtos"
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
	err := u.db.Preload("Roles").Preload("Departments").Model(&models.Users{}).Select("id, name, email, role_id, department_id, created_at, updated_at").Find(&users).Error
	// err := u.db.Find(&users).Error
	return users, err
}

func (u *UsersRepo) GetUserByID(id uint) (*models.Users, error) {
	var user models.Users
	err := u.db.Model(&models.Users{}).First(&user, id).Error
	return &user, err
}

func (u *UsersRepo) GetUsersByRole(roleID uint) (*[]models.Users, error) {
	var user []models.Users
	err := u.db.Model(&models.Users{}).Preload("Roles").Preload("Departments").Select("id, name, email, role_id, department_id, created_at, updated_at").Find(&user, "role_id = ?", roleID).Error
	return &user, err
}

func (u *UsersRepo) CreateUser(user *models.Users) error {
	return u.db.Create(user).Error
}

func (u *UsersRepo) UpdateUser(user *models.Users) error {
	if err := u.db.Model(user).Updates(user).Error; err != nil {
		return err
	}
	return nil
}

func (u *UsersRepo) DeleteUser(id uint) error {
	return u.db.Delete(&models.Users{}, id).Error
}

func (u *UsersRepo) LoginUser(user *models.Users) (*models.Users, error) {
	var users models.Users
	err := u.db.Preload("Roles").Preload("Departments").First(&users, "email = ?", user.Email).Error
	return &users, err
}

func (u *UsersRepo) SetPaymentMethodsToUser(request *dtos.UserPaymentMethodDTO) error {
	tx := u.db.Begin()

	var user models.Users
	if err := tx.Model(&models.Users{}).First(&user, request.UserID).Error; err != nil {
		tx.Rollback()
		return err
	}

	var paymentMethods []models.PaymentMethod
	if err := tx.Find(&paymentMethods, "code in (?)", request.PaymentMethods).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(paymentMethods) == 0 {
		tx.Rollback()
		return fmt.Errorf("payment method not found")
	}

	if err := tx.Model(&user).Association("PaymentMethods").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&user).Association("PaymentMethods").Append(paymentMethods); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (u *UsersRepo) GetUsersWithPaymentMethods() (*[]models.Users, error) {
	var users []models.Users
	err := u.db.Select("id, name, email, role_id, department_id").Joins("JOIN users_payment_methods ON users_payment_methods.users_id = users.id").Preload("PaymentMethods").Group("users.id").Find(&users).Error
	return &users, err
}

func (u *UsersRepo) GetUserPaymentMethods(userID uint) (*[]models.PaymentMethod, error) {
	var paymentMethods []models.PaymentMethod
	err := u.db.Joins("JOIN users_payment_methods ON users_payment_methods.payment_method_code = payment_methods.code").Where("users_payment_methods.users_id = ?", userID).Find(&paymentMethods).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethods, nil
}
