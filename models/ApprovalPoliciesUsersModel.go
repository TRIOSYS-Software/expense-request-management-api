package models

type ApprovalPoliciesUsers struct {
	ApprovalPolicyID uint `json:"approval_policy_id,omitempty" gorm:"primaryKey;"`
	UserID           uint `json:"user_id,omitempty" gorm:"primaryKey;"`
	Level            uint `json:"level,omitempty" gorm:"not null"`
	// ApprovalPolicy   ApprovalPolicies `gorm:"foreignKey:ApprovalPolicyID;references:ID"`
	Approver Users `gorm:"foreignKey:UserID;references:ID"`
}
