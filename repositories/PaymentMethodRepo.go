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

		delRes := tx.Where("CODE NOT IN ?", keep).Delete(&models.PaymentMethod{})
		if delRes.Error != nil {
			return delRes.Error
		}
		counts.Deleted = delRes.RowsAffected

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
