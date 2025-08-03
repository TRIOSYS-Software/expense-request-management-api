package controllers

import (
	"log"
	"net/http"
	"shwetaik-expense-management-api/services"

	"github.com/labstack/echo/v4"
)

type ProjectController struct {
	projectService *services.ProjectService
}

func NewProjectController(projectService *services.ProjectService) *ProjectController {
	return &ProjectController{projectService: projectService}
}

// SyncProjects synchronizes projects
// @Summary Sync projects
// @Description Synchronizes projects
// @Accept json
// @Produce json
// @Tags Project
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/sync [post]
// @Security JWT Token
func (c *ProjectController) SyncProjects(ctx echo.Context) error {
	err := c.projectService.SyncProjects()
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to sync projects",
		})
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"message": "Projects synced successfully",
	})
}

// GetProjects fetches projects
// @Summary Get projects
// @Description Fetches projects
// @Tags Project
// @Accept json
// @Produce json
// @Success 200 {object} []models.Project
// @Failure 500 {object} map[string]interface{}
// @Router /projects [get]
// @Security JWT Token
func (c *ProjectController) GetProjects(ctx echo.Context) error {
	projects, err := c.projectService.GetProjects()
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to fetch projects",
		})
	}
	return ctx.JSON(http.StatusOK, projects)
}
