package models

import "time"

type ExpenseApprovals struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	RequestID    uint       `json:"request_id" gorm:"not null"`
	ApproverID   uint       `json:"approver_id" gorm:"not null"`
	Level        uint       `json:"level" gorm:"not null"`
	Status       string     `json:"status" gorm:"not null"`
	Comments     *string    `json:"comments" gorm:"nullable"`
	ApprovalDate *time.Time `json:"approval_date" gorm:"nullable"`
}
