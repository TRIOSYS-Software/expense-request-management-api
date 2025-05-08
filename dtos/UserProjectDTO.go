package dtos

type UserProjectDTO struct {
	UserID   uint     `json:"user_id"`
	Projects []string `json:"projects"`
}
