package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type PaymentMethodRepo struct {
	db *gorm.DB
}

func NewPaymentMethodRepo(db *gorm.DB) *PaymentMethodRepo {
	return &PaymentMethodRepo{db: db}
}

func (r *PaymentMethodRepo) GetPaymentMethods() ([]models.PaymentMethod, error) {
	var paymentMethods []models.PaymentMethod
	err := r.db.Find(&paymentMethods).Error
	if err != nil {
		return nil, err
	}
	return paymentMethods, nil
}

func (r *PaymentMethodRepo) GetPaymentMethodByCode(code string) (*models.PaymentMethod, error) {
	var paymentMethod models.PaymentMethod
	err := r.db.Where("code = ?", code).First(&paymentMethod).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}

func (r *PaymentMethodRepo) SavePaymentMethods(paymentMethods []models.PaymentMethod) error {
	for _, paymentMethod := range paymentMethods {
		err := r.db.Save(&paymentMethod).Error
		if err != nil {
			return err
		}
	}
	return nil
}
