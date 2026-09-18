package controllers

import (
	"net/http"
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
	UploadDir              string
}

func NewExpenseRequestsController(expenseRequestsService *services.ExpenseRequestsService, uploadDir string) *ExpenseRequestsController {
	return &ExpenseRequestsController{
		ExpenseRequestsService: expenseRequestsService,
		UploadDir:              uploadDir,
	}
}

func parseOptionalFloat(c echo.Context, field string) *float64 {
	raw := strings.TrimSpace(c.FormValue(field))
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &v
}

func (er *ExpenseRequestsController) GetExpenseRequests(c echo.Context) error {
	var filterReq dtos.ExpenseRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Extract admin's user ID from JWT context (same behavior as approver)
	approverID, err := currentUserID(c)
	if err != nil {
		return unauthorized(c)
	}

	expenseRequests, total := er.ExpenseRequestsService.GetExpenseRequests(approverID, &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       expenseRequests,
		"pagination": pagination,
	})
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
func (er *ExpenseRequestsController) GetExpenseRequestByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid expense request id"})
	}
	expenseRequest, err := er.ExpenseRequestsService.GetExpenseRequestByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
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
func (er *ExpenseRequestsController) GetExpenseRequestsByUserID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user id"})
	}
	var filterReq dtos.ExpenseRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	expenseRequests, total := er.ExpenseRequestsService.GetExpenseRequestsByUserID(uint(i), &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       expenseRequests,
		"pagination": pagination,
	})
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
func (er *ExpenseRequestsController) GetExpenseRequestsSummary(c echo.Context) error {
	filters, err := parseSummaryFilters(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// category_id is expense-only, so it stays out of the shared parser.
	if s := c.QueryParam("category_id"); s != "" {
		categoryID, convErr := strconv.Atoi(s)
		if convErr != nil {
			return c.JSON(http.StatusBadRequest, "Invalid category ID")
		}
		filters["category_id"] = uint(categoryID)
	}

	summary, err := er.ExpenseRequestsService.GetExpenseRequestsSummary(filters)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, summary)
}

func (er *ExpenseRequestsController) GetAnalytics(c echo.Context) error {
	filters := make(map[string]any)

	if s := c.QueryParam("start_date"); s != "" {
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid start date")
		}
		filters["start_date"] = s
	}

	if s := c.QueryParam("end_date"); s != "" {
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid end date")
		}
		filters["end_date"] = s
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

	result, err := er.ExpenseRequestsService.GetAnalytics(filters)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
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
func (er *ExpenseRequestsController) CreateExpenseRequest(c echo.Context) error {
	expenseRequest := new(models.ExpenseRequests)
	if err := c.Bind(expenseRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	expenseRequest.AdvanceUsedAmount = parseOptionalFloat(c, "advance_used_amount")
	expenseRequest.ReturnedAmount = parseOptionalFloat(c, "returned_amount")

	legacyName, err := saveLegacyAttachment(c, er.UploadDir)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	expenseRequest.Attachment = legacyName

	for _, att := range saveMultiAttachments(c, er.UploadDir) {
		expenseRequest.Attachments = append(expenseRequest.Attachments, models.ExpenseRequestAttachments{
			FilePath: att.StoredName,
			FileName: att.OriginalName,
			FileType: att.ContentType,
		})
	}

	if err := er.ExpenseRequestsService.CreateExpenseRequest(expenseRequest); err != nil {
		names := make([]string, 0, len(expenseRequest.Attachments)+1)
		if expenseRequest.Attachment != nil {
			names = append(names, *expenseRequest.Attachment)
		}
		for _, att := range expenseRequest.Attachments {
			names = append(names, att.FilePath)
		}
		discardStoredFiles(er.UploadDir, names...)
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
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
func (er *ExpenseRequestsController) GetExpenseRequestByApproverID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	var filterReq dtos.ExpenseRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	expenseRequests, total := er.ExpenseRequestsService.GetExpenseRequestByApproverID(uint(i), &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       expenseRequests,
		"pagination": pagination,
	})
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
func (er *ExpenseRequestsController) SendExpenseRequestToSQLACC(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid expense request id"})
	}
	if err := er.ExpenseRequestsService.SendExpenseRequestToSQLACC(uint(id)); err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
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
func (er *ExpenseRequestsController) UpdateExpenseRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid expense request id"})
	}
	expenseRequest := new(models.ExpenseRequests)
	if err := c.Bind(expenseRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	// Advance-settlement amounts are nullable; parse explicitly so blank fields stay NULL.
	expenseRequest.AdvanceUsedAmount = parseOptionalFloat(c, "advance_used_amount")
	expenseRequest.ReturnedAmount = parseOptionalFloat(c, "returned_amount")

	legacyName, err := saveLegacyAttachment(c, er.UploadDir)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	// Unlike create, an absent upload leaves the existing attachment in place.
	if legacyName != nil {
		expenseRequest.Attachment = legacyName
	}

	expenseRequest.KeptAttachmentIDs, expenseRequest.KeepLegacyAttachment = attachmentRetention(c)

	for _, att := range saveMultiAttachments(c, er.UploadDir) {
		expenseRequest.Attachments = append(expenseRequest.Attachments, models.ExpenseRequestAttachments{
			FilePath: att.StoredName,
			FileName: att.OriginalName,
			FileType: att.ContentType,
		})
	}

	if err := er.ExpenseRequestsService.UpdateExpenseRequest(uint(id), expenseRequest); err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
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
func (er *ExpenseRequestsController) DeleteExpenseRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid expense request id"})
	}
	if err := er.ExpenseRequestsService.DeleteExpenseRequest(uint(id)); err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, "Expense request deleted successfully")
}

func (er *ExpenseRequestsController) SoftDeleteExpenseRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid expense request id"})
	}
	if err := er.ExpenseRequestsService.SoftDeleteExpenseRequest(uint(id)); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Expense request soft-deleted"})
}

func (er *ExpenseRequestsController) CompleteExpenseRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid expense request id"})
	}
	var body dtos.CompleteExpenseRequestDTO
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	actorUserID, err := currentUserID(c)
	if err != nil {
		return unauthorized(c)
	}
	if err := er.ExpenseRequestsService.CompleteExpenseRequest(uint(id), actorUserID, body.Comment); err != nil {
		return c.JSON(http.StatusConflict, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Expense request completed"})
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
func (er *ExpenseRequestsController) ServeExpenseRequestAttachment(c echo.Context) error {
	return serveUploadedFile(c, er.UploadDir, c.Param("filename"))
}
