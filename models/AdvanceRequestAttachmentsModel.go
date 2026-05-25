package models

import "time"

type AdvanceRequestAttachments struct {
	ID                uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	AdvanceRequestID  uint            `json:"advance_request_id" gorm:"not null"`
	FilePath          string          `json:"file_path" gorm:"not null"`
	FileName          string          `json:"file_name" gorm:"not null"`
	FileType          string          `json:"file_type"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	AdvanceRequest    AdvanceRequests `json:"-" gorm:"foreignKey:AdvanceRequestID"`
}
