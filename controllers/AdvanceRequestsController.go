package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type AdvanceRequestsController struct {
	AdvanceRequestsService *services.AdvanceRequestsService
	UploadDir              string
}

func NewAdvanceRequestsController(svc *services.AdvanceRequestsService) *AdvanceRequestsController {
	return &AdvanceRequestsController{
		AdvanceRequestsService: svc,
		UploadDir:              configs.Envs.UploadDir,
	}
}

func (ac *AdvanceRequestsController) GetAdvanceRequests(c echo.Context) error {
	var filterReq dtos.AdvanceRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	approverID := uint(c.Get("user_id").(float64))
	advanceRequests, total := ac.AdvanceRequestsService.GetAdvanceRequests(approverID, &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       advanceRequests,
		"pagination": pagination,
	})
}

func (ac *AdvanceRequestsController) GetAdvanceRequestByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	advanceRequest, err := ac.AdvanceRequestsService.GetAdvanceRequestByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, advanceRequest)
}

func (ac *AdvanceRequestsController) GetAdvanceRequestsByUserID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user id"})
	}
	var filterReq dtos.AdvanceRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	advanceRequests, total := ac.AdvanceRequestsService.GetAdvanceRequestsByUserID(uint(i), &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       advanceRequests,
		"pagination": pagination,
	})
}

func (ac *AdvanceRequestsController) GetAdvanceRequestByApproverID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	var filterReq dtos.AdvanceRequestFilterDTO
	if err := c.Bind(&filterReq); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	filterReq.ApproverID = uint(i)
	advanceRequests, total := ac.AdvanceRequestsService.GetAdvanceRequestByApproverID(uint(i), &filterReq)
	pagination := dtos.NewPaginationResponse(filterReq.Page, filterReq.Limit(), int(total))
	return c.JSON(http.StatusOK, map[string]any{
		"data":       advanceRequests,
		"pagination": pagination,
	})
}

func (ac *AdvanceRequestsController) GetAdvanceRequestsSummary(c echo.Context) error {
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
	if c.QueryParam("status") != "" {
		filters["status"] = c.QueryParam("status")
	}

	if v, err := strconv.ParseBool(c.QueryParam("need_my_approval")); err == nil && v {
		filters["need_my_approval"] = true
	}

	summary, err := ac.AdvanceRequestsService.GetAdvanceRequestsSummary(filters)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, summary)
}

func (ac *AdvanceRequestsController) GetSelectableAdvanceRequests(c echo.Context) error {
	userID := uint(c.Get("user_id").(float64))
	if uq := c.QueryParam("user_id"); uq != "" {
		if v, err := strconv.Atoi(uq); err == nil {
			userID = uint(v)
		}
	}
	list, err := ac.AdvanceRequestsService.GetSelectableAdvanceRequests(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, list)
}

func (ac *AdvanceRequestsController) CreateAdvanceRequest(c echo.Context) error {
	advanceRequest := new(models.AdvanceRequests)
	if err := c.Bind(advanceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
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

		if err := os.MkdirAll(ac.UploadDir, os.ModePerm); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create upload directory")
		}

		dstPath := filepath.Join(ac.UploadDir, uniqueFileName)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create file")
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to save file")
		}
		advanceRequest.Attachment = &uniqueFileName
	} else {
		advanceRequest.Attachment = nil
	}

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

			if err := os.MkdirAll(ac.UploadDir, os.ModePerm); err != nil {
				continue
			}

			dstPath := filepath.Join(ac.UploadDir, uniqueFileName)
			dst, err := os.Create(dstPath)
			if err != nil {
				continue
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				continue
			}

			advanceRequest.Attachments = append(advanceRequest.Attachments, models.AdvanceRequestAttachments{
				FilePath: uniqueFileName,
				FileName: file.Filename,
				FileType: file.Header.Get("Content-Type"),
			})
		}
	}

	if err := ac.AdvanceRequestsService.CreateAdvanceRequest(advanceRequest); err != nil {
		if advanceRequest.Attachment != nil {
			dstPath := filepath.Join(ac.UploadDir, *advanceRequest.Attachment)
			os.Remove(dstPath)
		}
		for _, att := range advanceRequest.Attachments {
			dstPath := filepath.Join(ac.UploadDir, att.FilePath)
			os.Remove(dstPath)
		}
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, advanceRequest)
}

func (ac *AdvanceRequestsController) UpdateAdvanceRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	advanceRequest := new(models.AdvanceRequests)
	if err := c.Bind(advanceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
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
		if err := os.MkdirAll(ac.UploadDir, os.ModePerm); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create upload directory")
		}
		dstPath := filepath.Join(ac.UploadDir, uniqueFileName)
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to create file")
		}
		defer dst.Close()
		if _, err := io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to save file")
		}
		advanceRequest.Attachment = &uniqueFileName
	}

	form, _ := c.MultipartForm()
	if form != nil {
		if keptIDs, ok := form.Value["kept_attachment_ids"]; ok {
			for _, idStr := range keptIDs {
				if v, err := strconv.Atoi(idStr); err == nil {
					advanceRequest.KeptAttachmentIDs = append(advanceRequest.KeptAttachmentIDs, uint(v))
				}
			}
		}
		if val, ok := form.Value["keep_legacy_attachment"]; ok && len(val) > 0 {
			advanceRequest.KeepLegacyAttachment, _ = strconv.ParseBool(val[0])
		}

		files := form.File["attachments"]
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				continue
			}
			defer src.Close()

			ext := filepath.Ext(file.Filename)
			uniqueFileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "multi", ext)
			if err := os.MkdirAll(ac.UploadDir, os.ModePerm); err != nil {
				continue
			}
			dstPath := filepath.Join(ac.UploadDir, uniqueFileName)
			dst, err := os.Create(dstPath)
			if err != nil {
				continue
			}
			defer dst.Close()
			if _, err := io.Copy(dst, src); err != nil {
				continue
			}
			advanceRequest.Attachments = append(advanceRequest.Attachments, models.AdvanceRequestAttachments{
				FilePath: uniqueFileName,
				FileName: file.Filename,
				FileType: file.Header.Get("Content-Type"),
			})
		}
	}

	if err := ac.AdvanceRequestsService.UpdateAdvanceRequest(uint(id), advanceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, advanceRequest)
}

func (ac *AdvanceRequestsController) DeleteAdvanceRequest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid advance request id"})
	}
	if err := ac.AdvanceRequestsService.DeleteAdvanceRequest(uint(id)); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, "Advance request deleted successfully")
}

func (ac *AdvanceRequestsController) ServeAdvanceRequestAttachment(c echo.Context) error {
	file := c.Param("filename")
	filePath := filepath.Join(ac.UploadDir, file)
	return c.File(filePath)
}
