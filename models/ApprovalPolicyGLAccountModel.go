package models

type ApprovalPolicyGLAccount struct {
	ApprovalPolicyID uint `gorm:"primaryKey;column:approval_policy_id"`
	GLAccountDockey  int  `gorm:"primaryKey;column:gl_account_dockey"`
}
