package models

import (
	"time"
)

type Users struct {
	ID           uint        `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Name         string      `json:"name" gorm:"not null"`
	Email        string      `json:"email" gorm:"unique;not null"`
	Password     string      `json:"password" gorm:"not null"`
	RoleID       uint        `json:"role" gorm:"not null"`
	DepartmentID uint        `json:"department" gorm:"not null"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt    time.Time   `json:"updated_at" gorm:"autoUpdateTime;not null"`
	Roles        Roles       `json:"roles" gorm:"foreignKey:RoleID;references:ID"`
	Departments  Departments `json:"departments" gorm:"foreignKey:DepartmentID;references:ID"`
}
