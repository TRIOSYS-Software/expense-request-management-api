package models

import "time"

type ExpenseApprovals struct {
	ID           uint       `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	RequestID    uint       `json:"request_id,omitempty" gorm:"not null;index:idx_expense_approvals_request_level,priority:1"`
	ApproverID   uint       `json:"approver_id,omitempty" gorm:"not null;index"`
	Level        uint       `json:"level,omitempty" gorm:"not null;index:idx_expense_approvals_request_level,priority:2"`
	Status       string     `json:"status,omitempty" gorm:"not null;type:enum('pending', 'approved', 'rejected')"`
	Comments     *string    `json:"comments" gorm:"nullable"`
	ApprovalDate *time.Time `json:"approval_date" gorm:"nullable"`
	IsFinal      bool       `json:"is_final" gorm:"not null;default:false"`
	Users        Users      `json:"users,omitempty" gorm:"foreignKey:ApproverID;references:ID"`
}
