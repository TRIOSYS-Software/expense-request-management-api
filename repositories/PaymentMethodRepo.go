package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *PaymentMethodRepo) GetPaymentMethodByCode(code string) (*models.PaymentMethod, error) {
	var paymentMethod models.PaymentMethod
	err := r.db.Where("code = ?", code).First(&paymentMethod).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}

// SavePaymentMethods upserts the supplied set and removes any locally cached
// rows whose CODE is not in the new set and is not still referenced by a
// request or a user assignment, all within one transaction.
func (r *PaymentMethodRepo) SavePaymentMethods(paymentMethods []models.PaymentMethod) (SyncCounts, error) {
	var counts SyncCounts
	if len(paymentMethods) == 0 {
		return counts, nil
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		keep := make([]string, 0, len(paymentMethods))
		for _, p := range paymentMethods {
			keep = append(keep, p.CODE)
		}

		deleted, retained, err := reconcile(tx, "payment_methods", "CODE", keep, paymentMethodRefs)
		if err != nil {
			return err
		}
		counts.Deleted = deleted
		counts.Retained = retained

		upRes := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "CODE"}},
			UpdateAll: true,
		}).Create(&paymentMethods)
		if upRes.Error != nil {
			return upRes.Error
		}
		counts.Upserted = upRes.RowsAffected
		return nil
	})
	return counts, err
}
