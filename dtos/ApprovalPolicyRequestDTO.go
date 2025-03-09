package dtos

type ApproverInput struct {
	ApproverID uint `json:"approver_id"`
	Level      uint `json:"level"`
}

type ApprovalPolicyRequestDTO struct {
	ConditionType  string          `json:"condition_type"`
	ConditionValue string          `json:"condition_value"`
	DepartmentID   *uint           `json:"department_id"`
	Priority       uint            `json:"priority"`
	Approvers      []ApproverInput `json:"approvers"`
}
