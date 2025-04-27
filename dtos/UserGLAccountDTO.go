package dtos

type UserGLAccountDTO struct {
	UserID     uint   `json:"user_id"`
	GLAccounts []uint `json:"gl_accounts"`
}
