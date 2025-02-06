package models

import "time"

type ExpenseRequests struct {
	ID            uint           `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Amount        float64        `json:"amount" gorm:"not null"`
	Reason        string         `json:"reason" gorm:"not null"`
	CategoryID    *uint          `json:"category" gorm:"nullable"`
	Project       *string        `json:"project" gorm:"nullable"`
	Approvers     *string        `json:"approvers" gorm:"nullable"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	DateSubmitted time.Time      `json:"date_submitted" gorm:"not null"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	ExpenseItems  []ExpenseItems `json:"expense_items" gorm:"foreignKey:RequestId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
