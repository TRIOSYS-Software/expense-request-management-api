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

var paymentMethodRefs = []childRef{
	{Table: "expense_requests", Column: "payment_method"},
	{Table: "advance_requests", Column: "payment_method"},
	{Table: "users_payment_methods", Column: "payment_method_code"},
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
	return replaceAll(r.db, paymentMethods, "payment_methods", "CODE", func(p models.PaymentMethod) string { return p.CODE }, paymentMethodRefs)
}
