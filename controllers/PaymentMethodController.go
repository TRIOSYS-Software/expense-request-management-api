package controllers

import (
	"log"
	"net/http"
	"shwetaik-expense-management-api/services"

	"github.com/labstack/echo/v4"
)

type PaymentMethodController struct {
	paymentMethodService *services.PaymentMethodService
}

func NewPaymentMethodController(paymentMethodService *services.PaymentMethodService) *PaymentMethodController {
	return &PaymentMethodController{paymentMethodService: paymentMethodService}
}

// SyncPaymentMethods synchronizes payment methods from the SQLACC API
// @Summary Sync payment methods
// @Description Synchronizes payment methods from the SQLACC API
// @Tags Payment Method
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Failure 500 {object} string
// @Router /payment-methods/sync [post]
// @Security JWT Token
func (c *PaymentMethodController) SyncPaymentMethods(ctx echo.Context) error {
	err := c.paymentMethodService.SyncPaymentMethods()
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, "Payment methods synced successfully")
}

// GetPaymentMethods fetches payment methods
// @Summary Get payment methods
// @Description Fetches payment methods
// @Tags Payment Method
// @Accept json
// @Produce json
// @Success 200 {object} []models.PaymentMethod
// @Failure 404 {object} string
// @Router /payment-methods [get]
// @Security JWT Token
func (c *PaymentMethodController) GetPaymentMethods(ctx echo.Context) error {
	paymentMethods, err := c.paymentMethodService.GetPaymentMethods()
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusNotFound, err)
	}
	return ctx.JSON(http.StatusOK, paymentMethods)
}
