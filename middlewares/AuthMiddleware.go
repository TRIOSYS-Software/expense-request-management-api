package middlewares

import (
	"log"
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
			log.Println("token not found")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		tsk, err := validateToken(token)
		if err != nil || !tsk.Valid {
			log.Println(err.Error())
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		claims, ok := tsk.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("invalid token")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		c.Set("user_id", claims["user_id"])
		c.Set("user_role", claims["user_role"])
		return next(c)
	}
}

func IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userRole := uint(c.Get("user_role").(float64))
		if userRole != 1 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return next(c)
	}
}
