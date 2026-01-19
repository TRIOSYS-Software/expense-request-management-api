package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type ExpenseRequestsController struct {
	ExpenseRequestsService *services.ExpenseRequestsService
	UploadDir              string
}

func NewExpenseRequestsController(expenseRequestsService *services.ExpenseRequestsService) *ExpenseRequestsController {
	return &ExpenseRequestsController{
		ExpenseRequestsService: expenseRequestsService,
		UploadDir:              configs.Envs.UploadDir,
	}
}

func (ex *ExpenseRequestsController) GetExpenseRequests(c echo.Context) error {
	expenseRequests := ex.ExpenseRequestsService.GetExpenseRequests()
	return c.JSON(http.StatusOK, expenseRequests)
}

// GetExpenseRequestByID returns a expense request by id
// @Summary Get a expense request by id
// @Description Get a expense request by id
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param id path int true "Expense request id"
// @Success 200 {object} models.ExpenseRequests
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/{id} [get]
// @Security JWT Token
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

// GetExpenseRequestsByUserID returns a expense request by user id
// @Summary Get a expense request by user id
// @Description Get a expense request by user id
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param id path int true "User id"
// @Success 200 {object} models.ExpenseRequests
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/user/{id} [get]
// @Security JWT Token
func (ex *ExpenseRequestsController) GetExpenseRequestsByUserID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	expenseRequests := ex.ExpenseRequestsService.GetExpenseRequestsByUserID(uint(i))
	return c.JSON(http.StatusOK, expenseRequests)
}

// GetExpenseRequestsSummary returns a expense request summary
// @Summary Get a expense request summary
// @Description Get a expense request summary
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Param category_id query int false "Category id"
// @Param user_id query int false "User id"
// @Param approver_id query int false "Approver id"
// @Param status query string false "Status"
// @Success 200 {object} dtos.ExpenseRequestSummary
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/summary [get]
// @Security JWT Token
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

// CreateExpenseRequest creates a new expense request
// @Summary Create a new expense request
// @Description Create a new expense request
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param ExpenseRequest body models.ExpenseRequests true "ExpenseRequest"
// @Success 200 {object} models.ExpenseRequests
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests [post]
// @Security JWT Token
func (ex *ExpenseRequestsController) CreateExpenseRequest(c echo.Context) error {
	expenseRequest := new(models.ExpenseRequests)
	if err := c.Bind(expenseRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	file, err := c.FormFile("attachment")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to open file")
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		uniqueFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		if err := os.MkdirAll(ex.UploadDir, os.ModePerm); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create upload directory")
		}

		dstPath := filepath.Join(ex.UploadDir, uniqueFileName)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create file")
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to save file")
		}
		expenseRequest.Attachment = &uniqueFileName
	} else {
		expenseRequest.Attachment = nil
	}

	// Handle Multiple Attachments
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["attachments"]
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				continue
			}
			defer src.Close()

			ext := filepath.Ext(file.Filename)
			uniqueFileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "multi", ext)

			if err := os.MkdirAll(ex.UploadDir, os.ModePerm); err != nil {
				continue
			}

			dstPath := filepath.Join(ex.UploadDir, uniqueFileName)
			dst, err := os.Create(dstPath)
			if err != nil {
				continue
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				continue
			}

			expenseRequest.Attachments = append(expenseRequest.Attachments, models.ExpenseRequestAttachments{
				FilePath: uniqueFileName,
				FileName: file.Filename,
				FileType: file.Header.Get("Content-Type"),
			})
		}
	}

	if err := ex.ExpenseRequestsService.CreateExpenseRequest(expenseRequest); err != nil {
		if expenseRequest.Attachment != nil {
			dstPath := filepath.Join(ex.UploadDir, *expenseRequest.Attachment)
			os.Remove(dstPath)
		}
		// Cleanup multi attachments on failure
		for _, att := range expenseRequest.Attachments {
			dstPath := filepath.Join(ex.UploadDir, att.FilePath)
			os.Remove(dstPath)
		}
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, expenseRequest)
}

// GetExpenseRequestByApproverID returns a list of expense requests by approver ID
// @Summary Get expense requests by approver ID
// @Description Get expense requests by approver ID
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param id path int true "Approver ID"
// @Success 200 {object} []models.ExpenseRequests
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/approvers/{id} [get]
// @Security JWT Token
func (ex *ExpenseRequestsController) GetExpenseRequestByApproverID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	expenseRequests := ex.ExpenseRequestsService.GetExpenseRequestByApproverID(uint(i))
	return c.JSON(http.StatusOK, expenseRequests)
}

// SendExpenseRequestToSQLACC sends an expense request to SQLACC
// @Summary Send an expense request to SQLACC
// @Description Send an expense request to SQLACC
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param id path int true "ExpenseRequest ID"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/{id}/sqlacc [post]
// @Security JWT Token
func (ex *ExpenseRequestsController) SendExpenseRequestToSQLACC(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid expense request id")
	}
	if err := ex.ExpenseRequestsService.SendExpenseRequestToSQLACC(uint(id)); err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, "Expense request sent to SQLACC successfully")
}

// UpdateExpenseRequest updates an expense request
// @Summary Update an expense request
// @Description Update an expense request
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param id path int true "ExpenseRequest ID"
// @Param ExpenseRequest body models.ExpenseRequests true "ExpenseRequest"
// @Success 200 {object} models.ExpenseRequests
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/{id} [put]
// @Security JWT Token
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
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to open file")
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		uniqueFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		if err := os.MkdirAll(ex.UploadDir, os.ModePerm); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create upload directory")
		}

		dstPath := filepath.Join(ex.UploadDir, uniqueFileName)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create file")
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to save file")
		}

		expenseRequest.Attachment = &uniqueFileName
	}

	// Manually bind KeptAttachmentIDs and KeepLegacyAttachment
	form, _ := c.MultipartForm()
	if form != nil {
		if keptIDs, ok := form.Value["kept_attachment_ids"]; ok {
			for _, idStr := range keptIDs {
				if id, err := strconv.Atoi(idStr); err == nil {
					expenseRequest.KeptAttachmentIDs = append(expenseRequest.KeptAttachmentIDs, uint(id))
				}
			}
		}
		if val, ok := form.Value["keep_legacy_attachment"]; ok && len(val) > 0 {
			expenseRequest.KeepLegacyAttachment, _ = strconv.ParseBool(val[0])
		}
	}

	// Handle Multiple Attachments for Update
	if form, err = c.MultipartForm(); err == nil {
		files := form.File["attachments"]
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				continue
			}
			defer src.Close()

			ext := filepath.Ext(file.Filename)
			uniqueFileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "multi", ext)

			if err := os.MkdirAll(ex.UploadDir, os.ModePerm); err != nil {
				continue
			}

			dstPath := filepath.Join(ex.UploadDir, uniqueFileName)
			dst, err := os.Create(dstPath)
			if err != nil {
				continue
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				continue
			}

			expenseRequest.Attachments = append(expenseRequest.Attachments, models.ExpenseRequestAttachments{
				FilePath: uniqueFileName,
				FileName: file.Filename,
				FileType: file.Header.Get("Content-Type"),
			})
		}
	}

	if err := ex.ExpenseRequestsService.UpdateExpenseRequest(uint(id), expenseRequest); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, expenseRequest)
}

// DeleteExpenseRequest deletes an expense request
// @Summary Delete an expense request
// @Description Delete an expense request
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param id path int true "ExpenseRequest ID"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/{id} [delete]
// @Security JWT Token
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

// ServeExpenseRequestAttachment serve expense request attachment
// @Summary Serve expense request attachment
// @Description Serve expense request attachment
// @Tags ExpenseRequests
// @Accept json
// @Produce json
// @Param filename path string true "Attachment filename"
// @Success 200 {file} file
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /expense-requests/attachment/{filename} [get]
// @Security JWT Token
func (ex *ExpenseRequestsController) ServeExpenseRequestAttachment(c echo.Context) error {
	file := c.Param("filename")
	filePath := filepath.Join(ex.UploadDir, file)
	return c.File(filePath)
}
