package models

import (
	"time"
)

type DeviceToken struct {
	ID        uint      `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	UserID    uint      `json:"user_id,omitempty" gorm:"not null;index"` 
	Token     string    `json:"token,omitempty" gorm:"unique;not null"`  
	DeviceOS  string    `json:"device_os,omitempty"`                      
	CreatedAt time.Time `json:"-" gorm:"autoCreateTime;not null"`
}