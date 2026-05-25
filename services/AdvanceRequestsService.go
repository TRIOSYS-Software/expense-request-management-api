package services

import (
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type AdvanceRequestsService struct {
	AdvanceRequestsRepo *repositories.AdvanceRequestsRepo
}

func NewAdvanceRequestsService(repo *repositories.AdvanceRequestsRepo) *AdvanceRequestsService {
	return &AdvanceRequestsService{AdvanceRequestsRepo: repo}
}

func (s *AdvanceRequestsService) GetAdvanceRequests(approverID uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	return s.AdvanceRequestsRepo.GetAdvanceRequests(approverID, filter)
}

func (s *AdvanceRequestsService) GetAdvanceRequestByID(id uint) (*models.AdvanceRequests, error) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestByID(id)
}

func (s *AdvanceRequestsService) GetAdvanceRequestsByUserID(id uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestsByUserID(id, filter)
}

func (s *AdvanceRequestsService) GetAdvanceRequestByApproverID(id uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestByApproverID(id, filter)
}

func (s *AdvanceRequestsService) GetAdvanceRequestsSummary(filters map[string]any) (dtos.AdvanceRequestSummary, error) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestsSummary(filters)
}

func (s *AdvanceRequestsService) GetSelectableAdvanceRequests(userID uint) ([]models.AdvanceRequests, error) {
	return s.AdvanceRequestsRepo.GetSelectableAdvanceRequests(userID)
}

func (s *AdvanceRequestsService) CreateAdvanceRequest(advanceRequest *models.AdvanceRequests) error {
	return s.AdvanceRequestsRepo.CreateAdvanceRequest(advanceRequest)
}

func (s *AdvanceRequestsService) UpdateAdvanceRequest(id uint, advanceRequest *models.AdvanceRequests) error {
	return s.AdvanceRequestsRepo.UpdateAdvanceRequest(id, advanceRequest)
}

func (s *AdvanceRequestsService) DeleteAdvanceRequest(id uint) error {
	return s.AdvanceRequestsRepo.DeleteAdvanceRequest(id)
}
