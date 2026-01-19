package models

import "time"

type ExpenseRequests struct {
	ID                   uint                        `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	Amount               float64                     `json:"amount,omitempty" form:"amount" gorm:"not null"`
	Description          string                      `json:"description,omitempty" form:"description" gorm:"not null"`
	Project              string                      `json:"project,omitempty" form:"project" gorm:"type:VARCHAR;size:20;not null"`
	PaymentMethod        string                      `json:"payment_method,omitempty" form:"payment_method" gorm:"type:VARCHAR;size:10;not null"`
	UserID               uint                        `json:"user_id,omitempty" form:"user_id" gorm:"not null"`
	GLAccount            string                      `json:"gl_account,omitempty" form:"gl_account" gorm:"not null"`
	DateSubmitted        time.Time                   `json:"date_submitted,omitempty" form:"date_submitted" gorm:"not null"`
	Attachment           *string                     `json:"attachment,omitempty" form:"attachment" gorm:"nullable"`
	CreatedAt            time.Time                   `json:"created_at,omitempty" gorm:"autoCreateTime;not null"`
	UpdatedAt            time.Time                   `json:"updated_at,omitempty" gorm:"autoUpdateTime;not null"`
	Status               string                      `json:"status,omitempty" gorm:"type:enum('pending', 'approved', 'rejected');not null;default:'pending'"`
	CurrentApproverLevel uint                        `json:"current_approver_level,omitempty" gorm:"not null;default:1"`
	IsSendToSQLACC       bool                        `json:"is_send_to_sql_acc" gorm:"not null;default:false"`
	Approvals            []ExpenseApprovals          `json:"approvals,omitempty" gorm:"foreignKey:RequestID"`
	User                 Users                       `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	PaymentMethods       PaymentMethod               `json:"payment_methods,omitempty" gorm:"foreignKey:PaymentMethod;references:CODE"`
	Projects             Project                     `json:"projects" gorm:"foreignKey:Project;reference:CODE"`
	GLAccounts           GLAcc                       `json:"gl_accounts,omitempty" gorm:"foreignKey:GLAccount;references:DOCKEY"`
	Attachments          []ExpenseRequestAttachments `json:"attachments,omitempty" gorm:"foreignKey:ExpenseRequestID"`
	KeptAttachmentIDs    []uint                      `json:"kept_attachment_ids,omitempty" form:"kept_attachment_ids" gorm:"-"`
	KeepLegacyAttachment bool                        `json:"keep_legacy_attachment,omitempty" form:"keep_legacy_attachment" gorm:"-"`
}
