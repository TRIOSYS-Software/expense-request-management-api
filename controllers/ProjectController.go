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
