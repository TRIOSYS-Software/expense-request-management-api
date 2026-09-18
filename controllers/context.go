package controllers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

var errNoClaim = errors.New("unauthorized")

func claimUint(c echo.Context, key string) (uint, error) {
	raw := c.Get(key)
	if raw == nil {
		return 0, errNoClaim
	}
	switch v := raw.(type) {
	case float64:
		if v < 0 {
			return 0, errNoClaim
		}
		return uint(v), nil
	case uint:
		return v, nil
	case int:
		if v < 0 {
			return 0, errNoClaim
		}
		return uint(v), nil
	default:
		return 0, errNoClaim
	}
}

// currentUserID returns the authenticated user's id.
func currentUserID(c echo.Context) (uint, error) { return claimUint(c, "user_id") }

// currentUserRoleID returns the authenticated user's role id.
func currentUserRoleID(c echo.Context) (uint, error) { return claimUint(c, "user_role") }

// unauthorized is the single 401 body shape these handlers return.
func unauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, echo.Map{"message": "unauthorized"})
}
