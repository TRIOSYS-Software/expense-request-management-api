package models

import "time"

type ExpenseRequestAttachments struct {
	ID               uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	ExpenseRequestID uint            `json:"expense_request_id" gorm:"not null"`
	FilePath         string          `json:"file_path" gorm:"not null"`
	FileName         string          `json:"file_name" gorm:"not null"`
	FileType         string          `json:"file_type"`
	CreatedAt        time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	ExpenseRequest   ExpenseRequests `json:"-" gorm:"foreignKey:ExpenseRequestID"`
}
