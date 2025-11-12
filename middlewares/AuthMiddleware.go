package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"shwetaik-expense-management-api/configs"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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

func RequirePermission(db *gorm.DB, entity string, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoleIDFloat := c.Get("user_role")

			if userRoleIDFloat == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}

			userRoleID, ok := userRoleIDFloat.(uint)
			if !ok {
				userRoleID = uint(c.Get("user_role").(float64))
			}

			var count int64
			err := db.
				Table("roles_permissions").
				Joins("JOIN permissions ON roles_permissions.permissions_id = permissions.id").
				Where("roles_permissions.roles_id = ? AND permissions.entity = ? AND permissions.action = ?", userRoleID, entity, action).
				Count(&count).Error

			fmt.Println(err)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Permission check failed"})
			}

			if count == 0 {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: Insufficient Permissions"})
			}

			return next(c)
		}
	}
}