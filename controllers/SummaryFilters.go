package controllers

import (
	"errors"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

const summaryDateLayout = "2006-01-02"

func parseSummaryFilters(c echo.Context) (map[string]any, error) {
	filters := make(map[string]any)

	if s := c.QueryParam("start_date"); s != "" {
		if _, err := time.Parse(summaryDateLayout, s); err != nil {
			return nil, errors.New("Invalid start date")
		}
		filters["start_date"] = s
	}

	if s := c.QueryParam("end_date"); s != "" {
		if _, err := time.Parse(summaryDateLayout, s); err != nil {
			return nil, errors.New("Invalid end date")
		}
		filters["end_date"] = s
	}

	if s := c.QueryParam("user_id"); s != "" {
		id, err := strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("Invalid user ID")
		}
		filters["user_id"] = uint(id)
	}

	if s := c.QueryParam("approver_id"); s != "" {
		id, err := strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("Invalid approver ID")
		}
		filters["approver_id"] = uint(id)
	}

	if s := c.QueryParam("status"); s != "" {
		filters["status"] = s
	}

	if v, err := strconv.ParseBool(c.QueryParam("need_my_approval")); err == nil && v {
		filters["need_my_approval"] = true
	}

	if s := c.QueryParam("search"); s != "" {
		filters["search"] = s
	}

	return filters, nil
}
