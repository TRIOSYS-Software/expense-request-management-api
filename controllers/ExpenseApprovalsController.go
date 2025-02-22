package controllers

import (
	"fmt"
	"net/http"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ExpenseApprovalsController struct {
	ExpenseApprovalsService *services.ExpenseApprovalsService
}

func NewExpenseApprovalsController(ExpenseApprovalsService *services.ExpenseApprovalsService) *ExpenseApprovalsController {
	return &ExpenseApprovalsController{ExpenseApprovalsService: ExpenseApprovalsService}
}

func (controller *ExpenseApprovalsController) GetExpenseApprovals(c echo.Context) error {
	ExpenseApprovals := controller.ExpenseApprovalsService.GetExpenseApprovals()
	return c.JSON(200, &ExpenseApprovals)
}

func (con *ExpenseApprovalsController) GetExpenseApprovalsByApproverID(c echo.Context) error {
	approverID := c.Param("approver_id")
	id, err := strconv.Atoi(approverID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	ExpenseApprovals := con.ExpenseApprovalsService.GetExpenseApprovalsByApproverID(uint(id))
	return c.JSON(200, &ExpenseApprovals)
}

func (con *ExpenseApprovalsController) UpdateExpenseApproval(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	ExpenseApprovals := new(models.ExpenseApprovals)
	if err := c.Bind(ExpenseApprovals); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	fmt.Println(ExpenseApprovals)
	if err := con.ExpenseApprovalsService.UpdateExpenseApproval(uint(idInt), ExpenseApprovals); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, ExpenseApprovals)
}
