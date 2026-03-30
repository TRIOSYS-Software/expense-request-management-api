package dtos

const (
	DEFAULT_PER_PAGE = 10
	MAX_PER_PAGE     = 100
)

type PaginationRequest struct {
	Page    int `json:"page" query:"page"`
	PerPage int `json:"per_page" query:"per_page"`
}

func (pr *PaginationRequest) Offset() int {
	if pr.Page <= 1 {
		return 0
	}
	return (pr.Page - 1) * pr.Limit()
}

func (pr *PaginationRequest) Limit() int {
	if pr.PerPage <= 0 {
		return DEFAULT_PER_PAGE
	}
	if pr.PerPage > MAX_PER_PAGE {
		return MAX_PER_PAGE
	}
	return pr.PerPage
}

type PaginationResponse struct {
	Total       int `json:"total"`
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
}

func NewPaginationResponse(page, perPage, total int) *PaginationResponse {
	if page <= 0 {
		page = 1
	}
	totalPages := (total + perPage - 1) / perPage
	return &PaginationResponse{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		TotalPages:  totalPages,
	}
}
