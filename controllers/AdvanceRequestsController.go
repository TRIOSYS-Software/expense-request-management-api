package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type AdvanceRequestsController struct {
	AdvanceRequestsService *services.AdvanceRequestsService
	UploadDir              string
}

func NewAdvanceRequestsController(svc *services.AdvanceRequestsService, uploadDir string) *AdvanceRequestsController {
	return &AdvanceRequestsController{
		AdvanceRequestsService: svc,
		UploadDir:              uploadDir,
	}
}

func (ar *AdvanceRequestsController) GetAdvanceRequests(c echo.Context) error {
	var filterReq dtos.AdvanceRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	approverID, err := currentUserID(c)
	if err != nil {
		return unauthorized(c)
	}
	advanceRequests, total := ar.AdvanceRequestsService.GetAdvanceRequests(approverID, &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       advanceRequests,
		"pagination": pagination,
	})
}

func (ar *AdvanceRequestsController) GetAdvanceRequestByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	advanceRequest, err := ar.AdvanceRequestsService.GetAdvanceRequestByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, advanceRequest)
}

func (ar *AdvanceRequestsController) GetAdvanceRequestsByUserID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user id"})
	}
	var filterReq dtos.AdvanceRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	advanceRequests, total := ar.AdvanceRequestsService.GetAdvanceRequestsByUserID(uint(i), &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       advanceRequests,
		"pagination": pagination,
	})
}

func (ar *AdvanceRequestsController) GetAdvanceRequestByApproverID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	var filterReq dtos.AdvanceRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	advanceRequests, total := ar.AdvanceRequestsService.GetAdvanceRequestByApproverID(uint(i), &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       advanceRequests,
		"pagination": pagination,
	})
}

func (ar *AdvanceRequestsController) GetAdvanceRequestsSummary(c echo.Context) error {
	filters, err := parseSummaryFilters(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	summary, err := ar.AdvanceRequestsService.GetAdvanceRequestsSummary(filters)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, summary)
}

func (ar *AdvanceRequestsController) GetSelectableAdvanceRequests(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return unauthorized(c)
	}
	if uq := c.QueryParam("user_id"); uq != "" {
		if v, err := strconv.Atoi(uq); err == nil {
			userID = uint(v)
		}
	}
	list, err := ar.AdvanceRequestsService.GetSelectableAdvanceRequests(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, list)
}

func (ar *AdvanceRequestsController) CreateAdvanceRequest(c echo.Context) error {
	advanceRequest := new(models.AdvanceRequests)
	if err := c.Bind(advanceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}

	legacyName, err := saveLegacyAttachment(c, ar.UploadDir)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	advanceRequest.Attachment = legacyName

	for _, att := range saveMultiAttachments(c, ar.UploadDir) {
		advanceRequest.Attachments = append(advanceRequest.Attachments, models.AdvanceRequestAttachments{
			FilePath: att.StoredName,
			FileName: att.OriginalName,
			FileType: att.ContentType,
		})
	}

	if err := ar.AdvanceRequestsService.CreateAdvanceRequest(advanceRequest); err != nil {
		names := make([]string, 0, len(advanceRequest.Attachments)+1)
		if advanceRequest.Attachment != nil {
			names = append(names, *advanceRequest.Attachment)
		}
		for _, att := range advanceRequest.Attachments {
			names = append(names, att.FilePath)
		}
		discardStoredFiles(ar.UploadDir, names...)
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, advanceRequest)
}

func (ar *AdvanceRequestsController) UpdateAdvanceRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	advanceRequest := new(models.AdvanceRequests)
	if err := c.Bind(advanceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}

	legacyName, err := saveLegacyAttachment(c, ar.UploadDir)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	// Unlike create, an absent upload leaves the existing attachment in place.
	if legacyName != nil {
		advanceRequest.Attachment = legacyName
	}

	advanceRequest.KeptAttachmentIDs, advanceRequest.KeepLegacyAttachment = attachmentRetention(c)

	for _, att := range saveMultiAttachments(c, ar.UploadDir) {
		advanceRequest.Attachments = append(advanceRequest.Attachments, models.AdvanceRequestAttachments{
			FilePath: att.StoredName,
			FileName: att.OriginalName,
			FileType: att.ContentType,
		})
	}

	if err := ar.AdvanceRequestsService.UpdateAdvanceRequest(uint(id), advanceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, advanceRequest)
}

func (ar *AdvanceRequestsController) DeleteAdvanceRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	if err := ar.AdvanceRequestsService.DeleteAdvanceRequest(uint(id)); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, "Advance request deleted successfully")
}

func (ar *AdvanceRequestsController) SoftDeleteAdvanceRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	linkedCount, _ := ar.AdvanceRequestsService.CountLinkedExpenseRequests(uint(id))
	if err := ar.AdvanceRequestsService.SoftDeleteAdvanceRequest(uint(id)); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message":                  "Advance request soft-deleted",
		"cascaded_expense_request": linkedCount,
	})
}

func (ar *AdvanceRequestsController) ServeAdvanceRequestAttachment(c echo.Context) error {
	return serveUploadedFile(c, ar.UploadDir, c.Param("filename"))
}

func (ar *AdvanceRequestsController) SendAdvanceRequestToSQLACC(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	if err := ar.AdvanceRequestsService.SendAdvanceRequestToSQLACC(uint(id)); err != nil {
		return c.JSON(http.StatusConflict, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, "Advance request sent to SQLACC successfully")
}

func (ar *AdvanceRequestsController) CloseAdvanceRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	var body dtos.CloseAdvanceRequestDTO
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	actorUserID, err := currentUserID(c)
	if err != nil {
		return unauthorized(c)
	}
	if err := ar.AdvanceRequestsService.CloseAdvanceRequest(uint(id), actorUserID, body.Comment); err != nil {
		return c.JSON(http.StatusConflict, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Advance request closed"})
}
