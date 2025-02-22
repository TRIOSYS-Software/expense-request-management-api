package middlewares

import (
	"net/http"
	"shwetaik-expense-management-api/configs"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func validateToken(token string) (*jwt.Token, error) {
	claims := jwt.MapClaims{}
	return jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(configs.Envs.JWTSecret), nil
	})
}

func IsAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		tsk, err := validateToken(token)
		if err != nil || !tsk.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		claims, ok := tsk.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		c.Set("user_id", claims["user_id"].(float64))
		return next(c)
	}
}

func IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		t, err := validateToken(token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok || claims["user_role"] != 1.0 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return next(c)
	}
}
