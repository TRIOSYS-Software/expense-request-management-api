package controllers

import (
	"log"
	"net/http"
	"shwetaik-expense-management-api/services"

	"github.com/labstack/echo/v4"
)

type GLAccController struct {
	GLAccService *services.GLAccService
}

func NewGLAccController(glAccService *services.GLAccService) *GLAccController {
	return &GLAccController{GLAccService: glAccService}
}

func (c *GLAccController) SyncGLAcc(ctx echo.Context) error {
	if err := c.GLAccService.SyncGLAcc(); err != nil {
		log.Printf("Error syncing GLAcc: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to sync GLAcc"})
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{"message": "GLAcc synced successfully"})
}

func (c *GLAccController) GetGLAcc(ctx echo.Context) error {
	glAcc, err := c.GLAccService.GetGLAcc()
	if err != nil {
		log.Printf("Error fetching GLAcc: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to fetch GLAcc"})
	}
	return ctx.JSON(http.StatusOK, glAcc)
}
