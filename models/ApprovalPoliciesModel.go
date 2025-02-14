package models

import "time"

type ApprovalPolicies struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	ConditionType  string    `json:"condition_type" gorm:"not null"`
	ConditionValue string    `json:"condition_value" gorm:"not null"`
	ApproverRoles  string    `json:"approver_roles" gorm:"not null"`
	DepartmentID   uint      `json:"department" gorm:"nullable"`
	Priority       uint      `json:"priority" gorm:"default:0"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
