package dtos

type PasswordResetRequestDTO struct {
	Email string `json:"email"`
}

type PasswordResetTokenDTO struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type PasswordResetChangeRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
