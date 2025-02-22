package models

import (
	"time"
)

type Users struct {
	ID               uint                `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	Name             string              `json:"name,omitempty" gorm:"not null"`
	Email            string              `json:"email,omitempty" gorm:"unique;not null"`
	Password         string              `json:"password,omitempty" gorm:"not null"`
	RoleID           uint                `json:"role,omitempty" gorm:"not null"`
	DepartmentID     uint                `json:"department,omitempty" gorm:"not null"`
	CreatedAt        time.Time           `json:"-" gorm:"autoCreateTime;not null"`
	UpdatedAt        time.Time           `json:"-" gorm:"autoUpdateTime;not null"`
	Roles            *Roles              `json:"roles,omitempty" gorm:"foreignKey:RoleID;references:ID"`
	Departments      *Departments        `json:"departments,omitempty" gorm:"foreignKey:DepartmentID;references:ID"`
	ApprovalPolicies []*ApprovalPolicies `gorm:"many2many:approval_policies_users;" json:"approval_policies,omitempty"`
}
