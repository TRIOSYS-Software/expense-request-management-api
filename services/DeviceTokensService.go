package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type DeviceTokenService struct {
	repo *repositories.DeviceTokenRepo
}

func NewDeviceTokenService(repo *repositories.DeviceTokenRepo) *DeviceTokenService {
	return &DeviceTokenService{repo: repo}
}

func (s *DeviceTokenService) GetTokensByUserID(userID uint) ([]string, error) {
	return s.repo.GetTokensByUserID(userID)
}

func (s *DeviceTokenService) CreateTokenByUserID(userID uint, token string, deviceOS string) (*models.DeviceToken, error) {
	return s.repo.CreateTokenByUserID(userID, token, deviceOS)
}
func (s *DeviceTokenService) DeleteToken(token string) error {
	return s.repo.DeleteToken(token)
}