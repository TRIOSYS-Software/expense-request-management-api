package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ApproversController struct {
	ApproversService *services.ApproversService
}

func NewApproversController(approversService *services.ApproversService) *ApproversController {
	return &ApproversController{ApproversService: approversService}
}

func (apr *ApproversController) GetApprovers(c echo.Context) error {
	users, err := apr.ApproversService.GetApprovers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (apr *ApproversController) GetApproverByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid approver id"})
	}
	user, err := apr.ApproversService.GetApproverByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func (apr *ApproversController) GetApproverActions(c echo.Context) error {
	approverID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid approver id"})
	}

	viewerIDRaw := c.Get("user_id")
	viewerRoleRaw := c.Get("user_role")
	if viewerIDRaw == nil || viewerRoleRaw == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "unauthorized"})
	}
	viewerID := uint(viewerIDRaw.(float64))
	viewerRoleID := uint(viewerRoleRaw.(float64))

	actions, err := apr.ApproversService.GetApproverActions(uint(approverID), viewerRoleID, viewerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, actions)
}
