package models

import "time"

type ApprovalPolicies struct {
	ID             uint                    `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	ConditionType  string                  `json:"condition_type,omitempty" gorm:"not null"`
	ConditionValue string                  `json:"condition_value,omitempty" gorm:"not null"`
	DepartmentID   *uint                   `json:"department,omitempty" gorm:"nullable"`
	Priority       uint                    `json:"priority,omitempty" gorm:"default:0"`
	CreatedAt      time.Time               `json:"created_at,omitempty" gorm:"autoCreateTime;not null"`
	UpdatedAt      time.Time               `json:"updated_at,omitempty" gorm:"autoUpdateTime;not null"`
	PolicyUsers    []ApprovalPoliciesUsers `json:"policy_users,omitempty" gorm:"foreignKey:ApprovalPolicyID;references:ID"`
	Departments    Departments             `json:"departments,omitempty" gorm:"foreignKey:DepartmentID;references:ID"`
}
