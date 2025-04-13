package models

import "time"

type ApprovalPolicies struct {
	ID           uint                    `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	MinAmount    float64                 `json:"min_amount" gorm:"not null"`
	MaxAmount    float64                 `json:"max_amount" gorm:"not null"`
	Project      string                  `json:"project" gorm:"not null"`
	DepartmentID *uint                   `json:"department" gorm:"nullable"`
	CreatedAt    time.Time               `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt    time.Time               `json:"updated_at" gorm:"autoUpdateTime;not null"`
	PolicyUsers  []ApprovalPoliciesUsers `json:"policy_users,omitempty" gorm:"foreignKey:ApprovalPolicyID;references:ID"`
	Departments  Departments             `json:"departments,omitempty" gorm:"foreignKey:DepartmentID;references:ID"`
	Projects     Project                 `json:"projects" gorm:"foreignKey:Project;reference:CODE"`
}
