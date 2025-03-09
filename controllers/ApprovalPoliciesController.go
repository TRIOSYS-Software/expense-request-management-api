package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ApprovalPoliciesController struct {
	approvalPoliciesService *services.ApprovalPoliciesService
}

func NewApprovalPoliciesController(approvalPoliciesService *services.ApprovalPoliciesService) *ApprovalPoliciesController {
	return &ApprovalPoliciesController{approvalPoliciesService: approvalPoliciesService}
}

func (u *ApprovalPoliciesController) GetApprovalPolicies(c echo.Context) error {
	approvalPolicies, err := u.approvalPoliciesService.GetApprovalPolicies()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, approvalPolicies)
}

func (u *ApprovalPoliciesController) GetApprovalPolicyByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid approval policy id")
	}
	approvalPolicy, err := u.approvalPoliciesService.GetApprovalPolicyByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, approvalPolicy)
}

func (u *ApprovalPoliciesController) CreateApprovalPolicy(c echo.Context) error {
	approvalPolicyDTO := new(dtos.ApprovalPolicyRequestDTO)
	if err := c.Bind(approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := u.approvalPoliciesService.CreateApprovalPolicy(approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, approvalPolicyDTO)
}

func (u *ApprovalPoliciesController) UpdateApprovalPolicy(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid approval policy id")
	}
	approvalPolicyDTO := new(dtos.ApprovalPolicyRequestDTO)
	if err := c.Bind(approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.approvalPoliciesService.UpdateApprovalPolicy(uint(id), approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, approvalPolicyDTO)
}

func (u *ApprovalPoliciesController) DeleteApprovalPolicy(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid approval policy id")
	}
	if err := u.approvalPoliciesService.DeleteApprovalPolicy(uint(i)); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, "Approval policy deleted successfully")
}
