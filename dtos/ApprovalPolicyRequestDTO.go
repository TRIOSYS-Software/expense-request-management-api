package dtos

type ApproverInput struct {
	ApproverID uint `json:"approver_id"`
	Level      uint `json:"level"`
}

type ApprovalPolicyRequestDTO struct {
	MinAmount    float64         `json:"min_amount"`
	MaxAmount    float64         `json:"max_amount"`
	Project      string          `json:"project"`
	DepartmentID *uint           `json:"department_id"`
	Approvers    []ApproverInput `json:"approvers"`
}
