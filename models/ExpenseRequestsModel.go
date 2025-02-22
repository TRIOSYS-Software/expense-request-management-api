package models

import "time"

type ExpenseRequests struct {
	ID                   uint               `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	Amount               float64            `json:"amount,omitempty" gorm:"not null"`
	Description          string             `json:"description,omitempty" gorm:"not null"`
	CategoryID           *uint              `json:"category_id,omitempty" gorm:"nullable"`
	Project              *string            `json:"project,omitempty" gorm:"nullable"`
	Approvers            *string            `json:"approvers,omitempty" gorm:"nullable"`
	UserID               uint               `json:"user_id,omitempty" gorm:"not null"`
	DateSubmitted        time.Time          `json:"date_submitted,omitempty" gorm:"not null"`
	CreatedAt            time.Time          `json:"created_at,omitempty" gorm:"autoCreateTime;not null"`
	UpdatedAt            time.Time          `json:"updated_at,omitempty" gorm:"autoUpdateTime;not null"`
	Status               string             `json:"status,omitempty" gorm:"type:enum('pending', 'approved', 'rejected');not null;default:'pending'"`
	CurrentApproverLevel uint               `json:"current_approver_level,omitempty" gorm:"not null;default:1"`
	Approvals            []ExpenseApprovals `json:"approvals,omitempty" gorm:"foreignKey:RequestID"`
	Category             ExpenseCategories  `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
	User                 Users              `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}
