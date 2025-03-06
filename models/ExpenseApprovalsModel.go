package models

import "time"

type ExpenseApprovals struct {
	ID           uint       `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	RequestID    uint       `json:"request_id,omitempty" gorm:"not null"`
	ApproverID   uint       `json:"approver_id,omitempty" gorm:"not null"`
	Level        uint       `json:"level,omitempty" gorm:"not null"`
	Status       string     `json:"status,omitempty" gorm:"not null;type:enum('pending', 'approved', 'rejected')"`
	Comments     *string    `json:"comments" gorm:"nullable"`
	ApprovalDate *time.Time `json:"approval_date" gorm:"nullable"`
	IsFinal      bool       `json:"is_final" gorm:"not null;default:false"`
	// ExpenseRequests ExpenseRequests `json:"expense_requests,omitempty" gorm:"foreignKey:RequestID;references:ID"`
	Users Users `json:"users,omitempty" gorm:"foreignKey:ApproverID;references:ID"`
}
