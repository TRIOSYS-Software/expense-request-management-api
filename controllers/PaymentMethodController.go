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

func (c *PaymentMethodController) SyncPaymentMethods(ctx echo.Context) error {
	err := c.paymentMethodService.SyncPaymentMethods()
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, "Payment methods synced successfully")
}

func (c *PaymentMethodController) GetPaymentMethods(ctx echo.Context) error {
	paymentMethods, err := c.paymentMethodService.GetPaymentMethods()
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusNotFound, err)
	}
	return ctx.JSON(http.StatusOK, paymentMethods)
}
