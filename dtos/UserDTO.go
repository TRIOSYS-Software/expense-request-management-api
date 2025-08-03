package dtos

import "shwetaik-expense-management-api/models"

type LoginRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponseDTO struct {
	User  models.Users `json:"User"`
	Token string       `json:"Token"`
}

type UserRequestDTO struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       uint   `json:"role"`
	Department *uint  `json:"department"`
}
