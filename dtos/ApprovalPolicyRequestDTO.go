package dtos

type ApproverInput struct {
	ApproverID uint `json:"approver_id"`
	Level      uint `json:"level"`
}

type ApprovalPolicyRequestDTO struct {
	PolicyType   string          `json:"policy_type"`
	MinAmount    float64         `json:"min_amount"`
	MaxAmount    float64         `json:"max_amount"`
	Project      string          `json:"project"`
	DepartmentID *uint           `json:"department_id"`
	GLAccountIDs []string        `json:"gl_account_ids"`
	Approvers    []ApproverInput `json:"approvers"`
}
