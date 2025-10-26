package dtos

type DeviceTokenRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Token    string `json:"token" binding:"required"`
	DeviceOS string `json:"device_os" binding:"required"`
}

type GetTokensByUserIDRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

type DeviceTokensResponse struct {
	UserID uint     `json:"user_id"`
	Tokens []string `json:"tokens"`
}