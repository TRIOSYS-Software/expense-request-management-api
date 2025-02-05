package models

import "time"

type ExpenseItems struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Amount      float64   `json:"amount" gorm:"not null"`
	Description string    `json:"reason" gorm:"not null"`
	RequestId   uint      `json:"request_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
