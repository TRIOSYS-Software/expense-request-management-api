package dtos

type ExpenseRequestFilterDTO struct {
	PaginationRequest
	Status             string   `json:"status" query:"status"`
	Search             string   `json:"search" query:"search"`
	StartDate          string   `json:"start_date" query:"start_date"`
	EndDate            string   `json:"end_date" query:"end_date"`
	IncludedAsApprover bool     `json:"included_as_approver" query:"included_as_approver"`
	NeedMyApproval     bool     `json:"need_my_approval" query:"need_my_approval"`
	ApproverID         uint     `json:"approver_id" query:"approver_id"`
	MinAmount          *float64 `json:"min_amount" query:"min_amount"`
	MaxAmount          *float64 `json:"max_amount" query:"max_amount"`
}
