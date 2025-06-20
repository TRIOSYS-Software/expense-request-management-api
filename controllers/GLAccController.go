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

// SyncGLAcc synchonizes GLAcc data from SQLAcc API
// @Summary Sync GLAcc
// @Description Sync GLAcc data from SQLAcc API
// @Tags GLAcc
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /gl-acc/sync [post]
// @Security JWT Token
func (c *GLAccController) SyncGLAcc(ctx echo.Context) error {
	if err := c.GLAccService.SyncGLAcc(); err != nil {
		log.Printf("Error syncing GLAcc: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to sync GLAcc"})
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{"message": "GLAcc synced successfully"})
}

// GetGLAcc fetches GLAcc data
// @Summary Get GLAcc
// @Description Get GLAcc data
// @Tags GLAcc
// @Accept json
// @Produce json
// @Success 200 {object} []models.GLAcc
// @Failure 500 {object} map[string]interface{}
// @Router /gl-acc [get]
// @Security JWT Token
func (c *GLAccController) GetGLAcc(ctx echo.Context) error {
	glAcc, err := c.GLAccService.GetGLAcc()
	if err != nil {
		log.Printf("Error fetching GLAcc: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to fetch GLAcc"})
	}
	return ctx.JSON(http.StatusOK, glAcc)
}
