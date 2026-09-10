package dtos

// UpdateExpenseApprovalCommentDTO is the body of PUT /approvals/{id}/comment.
type UpdateExpenseApprovalCommentDTO struct {
	Comments string `json:"comments"`
}

// UpdateAdvanceApprovalCommentDTO is the body of PUT /advance-approvals/{id}/comment.
type UpdateAdvanceApprovalCommentDTO struct {
	Comments string `json:"comments"`
}
