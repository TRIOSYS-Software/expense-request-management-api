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

func (u *UsersRepo) GetUserByEmail(email string) (*models.Users, error) {
	var user models.Users
	err := u.db.Model(&models.Users{}).First(&user, "email = ?", email).Error
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

	if err := tx.Model(&user).Association("PaymentMethods").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	if len(paymentMethods) != 0 {
		if err := tx.Model(&user).Association("PaymentMethods").Append(paymentMethods); err != nil {
			tx.Rollback()
			return err
		}
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

func (u *UsersRepo) SetGLAccountsToUser(request *dtos.UserGLAccountDTO) error {
	tx := u.db.Begin()

	var user models.Users
	if err := tx.Model(&models.Users{}).First(&user, request.UserID).Error; err != nil {
		tx.Rollback()
		return err
	}

	var glAccounts []models.GLAcc
	if err := tx.Find(&glAccounts, "DOCKEY in (?)", request.GLAccounts).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&user).Association("GLAccounts").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	if len(glAccounts) != 0 {
		if err := tx.Model(&user).Association("GLAccounts").Append(glAccounts); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (u *UsersRepo) GetUsersWithGLAccounts() (*[]models.Users, error) {
	var users []models.Users
	err := u.db.Select("id, name, email, role_id, department_id").Joins("JOIN users_gl_accounts ON users_gl_accounts.users_id = users.id").Preload("GLAccounts").Group("users.id").Find(&users).Error
	return &users, err
}

func (u *UsersRepo) GetUserGLAccounts(userID uint) (*[]models.GLAcc, error) {
	var glAccounts []models.GLAcc
	err := u.db.Joins("JOIN users_gl_accounts ON users_gl_accounts.gl_acc_dockey = gl_accs.dockey").Where("users_gl_accounts.users_id = ?", userID).Find(&glAccounts).Error
	if err != nil {
		return nil, err
	}
	return &glAccounts, nil
}

func (u *UsersRepo) CreatePasswordReset(passwordReset *models.PasswordReset) error {
	return u.db.Create(passwordReset).Error
}

func (u *UsersRepo) ValidatePasswordResetToken(passwordReset *models.PasswordReset, token dtos.PasswordResetTokenDTO) error {
	user, err := u.GetUserByEmail(token.Email)
	if err != nil {
		return err
	}
	err = u.db.First(&passwordReset, "token = ? and user_id = ?", token.Token, user.ID).Error
	return err
}

func (u *UsersRepo) UpdatePasswordReset(passwordReset *models.PasswordReset) error {
	if err := u.db.Model(passwordReset).Updates(passwordReset).Error; err != nil {
		return err
	}
	fmt.Println(passwordReset)
	return nil
}

func (u *UsersRepo) DeletePasswordReset(passwordReset *models.PasswordReset) error {
	return u.db.Delete(&models.PasswordReset{}, "user_id = ?", passwordReset.UserID).Error
}

func (u *UsersRepo) SetProjectsToUser(request *dtos.UserProjectDTO) error {
	tx := u.db.Begin()

	var user models.Users
	if err := tx.Model(&models.Users{}).First(&user, request.UserID).Error; err != nil {
		tx.Rollback()
		return err
	}

	var projects []models.Project
	if err := tx.Find(&projects, "code in (?)", request.Projects).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&user).Association("Projects").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	if len(projects) != 0 {
		if err := tx.Model(&user).Association("Projects").Append(projects); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (u *UsersRepo) GetUsersWithProjects() (*[]models.Users, error) {
	var users []models.Users
	err := u.db.Select("id, name, email, role_id, department_id").Joins("JOIN users_projects ON users_projects.users_id = users.id").Preload("Projects").Group("users.id").Find(&users).Error
	return &users, err
}

func (u *UsersRepo) GetUserProjects(userID uint) (*[]models.Project, error) {
	var projects []models.Project
	err := u.db.Joins("JOIN users_projects ON users_projects.project_code = projects.code").Where("users_projects.users_id = ?", userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return &projects, nil
}
