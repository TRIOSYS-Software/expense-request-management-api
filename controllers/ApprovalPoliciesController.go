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

func (ap *ApprovalPoliciesController) GetApprovalPolicies(c echo.Context) error {
	policyType := c.QueryParam("policy_type")
	if policyType != "" && policyType != "expense" && policyType != "advance" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid policy_type"})
	}
	approvalPolicies, err := ap.approvalPoliciesService.GetApprovalPolicies(policyType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, approvalPolicies)
}

func (ap *ApprovalPoliciesController) GetApprovalPolicyByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid approval policy id"})
	}
	approvalPolicy, err := ap.approvalPoliciesService.GetApprovalPolicyByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, approvalPolicy)
}

func (ap *ApprovalPoliciesController) CreateApprovalPolicy(c echo.Context) error {
	approvalPolicyDTO := new(dtos.ApprovalPolicyRequestDTO)
	if err := c.Bind(approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}

	if err := ap.approvalPoliciesService.CreateApprovalPolicy(approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, approvalPolicyDTO)
}

func (ap *ApprovalPoliciesController) UpdateApprovalPolicy(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid approval policy id"})
	}
	approvalPolicyDTO := new(dtos.ApprovalPolicyRequestDTO)
	if err := c.Bind(approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	if err := ap.approvalPoliciesService.UpdateApprovalPolicy(uint(id), approvalPolicyDTO); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, approvalPolicyDTO)
}

func (ap *ApprovalPoliciesController) DeleteApprovalPolicy(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid approval policy id"})
	}
	if err := ap.approvalPoliciesService.DeleteApprovalPolicy(uint(i)); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, "Approval policy deleted successfully")
}
