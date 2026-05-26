package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type AdvanceApprovalsController struct {
	AdvanceApprovalsService *services.AdvanceApprovalsService
}

func NewAdvanceApprovalsController(svc *services.AdvanceApprovalsService) *AdvanceApprovalsController {
	return &AdvanceApprovalsController{AdvanceApprovalsService: svc}
}

func (c *AdvanceApprovalsController) GetAdvanceApprovals(ctx echo.Context) error {
	list := c.AdvanceApprovalsService.GetAdvanceApprovals()
	return ctx.JSON(http.StatusOK, &list)
}

func (c *AdvanceApprovalsController) GetAdvanceApprovalsByApproverID(ctx echo.Context) error {
	approverID := ctx.Param("approver_id")
	id, err := strconv.Atoi(approverID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	list := c.AdvanceApprovalsService.GetAdvanceApprovalsByApproverID(uint(id))
	return ctx.JSON(http.StatusOK, &list)
}

func (c *AdvanceApprovalsController) UpdateAdvanceApproval(ctx echo.Context) error {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	advanceApprovals := new(models.AdvanceApprovals)
	if err := ctx.Bind(advanceApprovals); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	if err := c.AdvanceApprovalsService.UpdateAdvanceApproval(uint(idInt), advanceApprovals); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return ctx.JSON(http.StatusOK, advanceApprovals)
}

type UpdateAdvanceApprovalCommentDTO struct {
	Comments string `json:"comments"`
}

func (c *AdvanceApprovalsController) UpdateAdvanceApprovalComment(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	var dto UpdateAdvanceApprovalCommentDTO
	if err := ctx.Bind(&dto); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	if err := c.AdvanceApprovalsService.UpdateAdvanceApprovalComment(uint(id), dto.Comments); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return ctx.JSON(http.StatusOK, echo.Map{"message": "Comment updated successfully"})
}
