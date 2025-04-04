package dtos

type UserPaymentMethodDTO struct {
	UserID         uint     `json:"user_id"`
	PaymentMethods []string `json:"payment_methods"`
}
