package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type DeviceTokenRepo struct {
	db *gorm.DB
}

func NewDeviceTokenRepo(db *gorm.DB) *DeviceTokenRepo {
	return &DeviceTokenRepo{db: db}
}

func (r *DeviceTokenRepo) GetTokensByUserID(userID uint) ([]string, error) {
	var tokens []string
	err := r.db.Model(&models.DeviceToken{}).Where("user_id = ?", userID).Pluck("token", &tokens).Error
	return tokens, err
}

func (r *DeviceTokenRepo) CreateTokenByUserID(userID uint, token string, deviceOS string) (*models.DeviceToken, error) {
	deviceToken := &models.DeviceToken{
		UserID:   userID,
		Token:    token,
		DeviceOS: deviceOS,
	}

	if err := r.db.Create(deviceToken).Error; err != nil {
		return nil, err
	}

	return deviceToken, nil
}