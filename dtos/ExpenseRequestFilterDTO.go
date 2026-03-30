package dtos

type ExpenseRequestFilterDTO struct {
	PaginationRequest
	Status       string `json:"status" query:"status"`
	Search       string `json:"search" query:"search"`
	Date         string `json:"date" query:"date"`
	ApprovedByMe bool   `json:"approved_by_me" query:"approved_by_me"`
	ApproverID   uint   `json:"approver_id" query:"approver_id"`
}
