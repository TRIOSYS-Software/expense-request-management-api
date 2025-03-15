package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type ExpenseRequestsController struct {
	ExpenseRequestsService *services.ExpenseRequestsService
}

func NewExpenseRequestsController(expenseRequestsService *services.ExpenseRequestsService) *ExpenseRequestsController {
	return &ExpenseRequestsController{ExpenseRequestsService: expenseRequestsService}
}

func (ex *ExpenseRequestsController) GetExpenseRequests(c echo.Context) error {
	expenseRequests := ex.ExpenseRequestsService.GetExpenseRequests()
	return c.JSON(http.StatusOK, expenseRequests)
}

func (ex *ExpenseRequestsController) GetExpenseRequestByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid expense request id")
	}
	expenseRequest, err := ex.ExpenseRequestsService.GetExpenseRequestByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, expenseRequest)
}

func (ex *ExpenseRequestsController) GetExpenseRequestsByUserID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	expenseRequests := ex.ExpenseRequestsService.GetExpenseRequestsByUserID(uint(i))
	return c.JSON(http.StatusOK, expenseRequests)
}

func (ex *ExpenseRequestsController) GetExpenseRequestsSummary(c echo.Context) error {
	filters := make(map[string]any)
	if c.QueryParam("start_date") != "" {
		startDate, err := time.Parse("2006-01-02", c.QueryParam("start_date"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid start date")
		}
		filters["start_date"] = startDate
	}

	if c.QueryParam("end_date") != "" {
		endDate, err := time.Parse("2006-01-02", c.QueryParam("end_date"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid end date")
		}
		filters["end_date"] = endDate
	}

	if c.QueryParam("category_id") != "" {
		categoryID, err := strconv.Atoi(c.QueryParam("category_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid category ID")
		}
		filters["category_id"] = uint(categoryID)
	}

	if c.QueryParam("user_id") != "" {
		userID, err := strconv.Atoi(c.QueryParam("user_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid user ID")
		}
		filters["user_id"] = uint(userID)
	}

	if c.QueryParam("approver_id") != "" {
		approverID, err := strconv.Atoi(c.QueryParam("approver_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid approver ID")
		}
		filters["approver_id"] = uint(approverID)
	}

	if c.QueryParam("status") != "" {
		status := c.QueryParam("status")
		filters["status"] = status
	}

	summary, err := ex.ExpenseRequestsService.GetExpenseRequestsSummary(filters)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, summary)
}

func (ex *ExpenseRequestsController) CreateExpenseRequest(c echo.Context) error {
	expenseRequest := new(models.ExpenseRequests)
	if err := c.Bind(expenseRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	file, err := c.FormFile("attachment")
	if err == nil {
		src, _ := file.Open()
		defer src.Close()

		filename := strings.Split(file.Filename, ".")

		uniqueFileName := fmt.Sprintf("%d.%s", time.Now().UnixNano(), filename[1])
		dstPath := filepath.Join("uploads", uniqueFileName)
		fmt.Println(dstPath)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create path")
		}

		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create file")
		}
		expenseRequest.Attachment = &uniqueFileName
	} else {
		expenseRequest.Attachment = nil
	}

	if err := ex.ExpenseRequestsService.CreateExpenseRequest(expenseRequest); err != nil {
		dstPath := filepath.Join("uploads", *expenseRequest.Attachment)
		os.Remove(dstPath)
		return c.JSON(http.StatusNotFound, err.Error())
	}
	fmt.Println(expenseRequest)
	return c.JSON(http.StatusOK, expenseRequest)
}

func (ex *ExpenseRequestsController) GetExpenseRequestByApproverID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	expenseRequests := ex.ExpenseRequestsService.GetExpenseRequestByApproverID(uint(i))
	return c.JSON(http.StatusOK, expenseRequests)
}

func (ex *ExpenseRequestsController) SendExpenseRequestToSQLACC(c echo.Context) error {
	expenseRequestDTO := new(dtos.ApprovedExpenseRequestsDTO)
	if err := c.Bind(expenseRequestDTO); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request payload")
	}
	fmt.Println(expenseRequestDTO)
	if err := ex.ExpenseRequestsService.SendExpenseRequestToSQLACC(expenseRequestDTO); err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, expenseRequestDTO)
}

func (ex *ExpenseRequestsController) UpdateExpenseRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid expense request id")
	}
	expenseRequest := new(models.ExpenseRequests)
	if err := c.Bind(expenseRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	file, err := c.FormFile("attachment")
	if err == nil {
		src, _ := file.Open()
		defer src.Close()

		filename := strings.Split(file.Filename, ".")

		uniqueFileName := fmt.Sprintf("%d.%s", time.Now().UnixNano(), filename[1])
		dstPath := filepath.Join("uploads", uniqueFileName)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create folder")
		}

		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create file")
		}

		expenseRequest.Attachment = &uniqueFileName
	}

	if err := ex.ExpenseRequestsService.UpdateExpenseRequest(uint(id), expenseRequest); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, expenseRequest)
}

func (ex *ExpenseRequestsController) DeleteExpenseRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid expense request id")
	}
	if err := ex.ExpenseRequestsService.DeleteExpenseRequest(uint(id)); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, "Expense request deleted successfully")
}

func (ex *ExpenseRequestsController) ServeExpenseRequestAttachment(c echo.Context) error {
	file := c.Param("filename")
	fmt.Println(file)
	filePath := filepath.Join("uploads", file)
	return c.File(filePath)
}
