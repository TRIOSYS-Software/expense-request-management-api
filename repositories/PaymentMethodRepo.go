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

func (r *PaymentMethodRepo) SavePaymentMethods(paymentMethods []models.PaymentMethod) (SyncCounts, error) {
	return replaceAll(r.db, paymentMethods, "CODE", func(p models.PaymentMethod) string { return p.CODE })
}
