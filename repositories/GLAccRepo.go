package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type GLAccRepo struct {
	db *gorm.DB
}

func NewGLAccRepo(db *gorm.DB) *GLAccRepo {
	return &GLAccRepo{db: db}
}

func (r *GLAccRepo) GetGLAcc() ([]models.GLAcc, error) {
	var glAcc []models.GLAcc
	err := r.db.Find(&glAcc).Error
	if err != nil {
		return nil, err
	}
	return glAcc, nil
}

func (r *GLAccRepo) ReplaceGLAcc(glAccs []models.GLAcc) (SyncCounts, error) {
	return replaceAll(r.db, glAccs, "DOCKEY", func(a models.GLAcc) int { return a.DOCKEY })
}
