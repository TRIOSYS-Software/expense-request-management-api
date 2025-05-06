package controllers

import (
	"log"
	"net/http"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type UsersController struct {
	UsersService *services.UsersService
}

func NewUsersController(usersService *services.UsersService) *UsersController {
	return &UsersController{UsersService: usersService}
}

func (u *UsersController) GetUsers(c echo.Context) error {
	users, err := u.UsersService.GetUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}

func (u *UsersController) GetUsersByRole(c echo.Context) error {
	roleID := c.Param("role_id")
	i, err := strconv.Atoi(roleID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid role id")
	}
	user, err := u.UsersService.GetUsersByRole(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UsersController) LoginUser(c echo.Context) error {
	user := new(models.Users)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	data, err := u.UsersService.LoginUser(user)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}
	return c.JSON(http.StatusOK, data)
}

func (u *UsersController) CreateUser(c echo.Context) error {
	user := new(models.Users)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.CreateUser(user); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UsersController) GetUserByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	user, err := u.UsersService.GetUserByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UsersController) UpdateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	user := new(models.Users)
	user.ID = uint(id)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.UpdateUser(user); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UsersController) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	if err := u.UsersService.DeleteUser(uint(i)); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, "User deleted successfully")
}

func (u *UsersController) VerifyUser(c echo.Context) error {
	user_id := c.Get("user_id")
	user, err := u.UsersService.GetUserByID(uint(user_id.(float64)))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid User!")
	}
	return c.JSON(http.StatusOK, user)
}

func (u *UsersController) SetPaymentMethodsToUser(c echo.Context) error {
	userPaymentMethodDTO := new(dtos.UserPaymentMethodDTO)
	if err := c.Bind(userPaymentMethodDTO); err != nil {
		log.Printf("Error binding payment methods to user: %v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.SetPaymentMethodsToUser(userPaymentMethodDTO); err != nil {
		log.Printf("Error setting payment methods to user: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, "Payment methods set successfully")
}

func (u *UsersController) GetUsersWithPaymentMethods(c echo.Context) error {
	users, err := u.UsersService.GetUsersWithPaymentMethods()
	if err != nil {
		log.Printf("Error getting users with payment methods: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}

func (u *UsersController) GetPaymentMethodsByUserID(ctx echo.Context) error {
	userID := ctx.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid user id")
	}
	paymentMethods, err := u.UsersService.GetUserPaymentMethods(uint(id))
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusNotFound, err)
	}
	return ctx.JSON(http.StatusOK, paymentMethods)
}

func (u *UsersController) SetGLAccountsToUser(c echo.Context) error {
	userGLAccountDTO := new(dtos.UserGLAccountDTO)
	if err := c.Bind(userGLAccountDTO); err != nil {
		log.Printf("Error binding GLAccounts to user: %v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.SetGLAccountsToUser(userGLAccountDTO); err != nil {
		log.Printf("Error setting GLAccounts to user: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, "GLAccounts set successfully")
}

func (u *UsersController) GetGLAccountsByUserID(ctx echo.Context) error {
	userID := ctx.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid user id")
	}
	GLAccounts, err := u.UsersService.GetUserGLAccounts(uint(id))
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusNotFound, err)
	}
	return ctx.JSON(http.StatusOK, GLAccounts)
}

func (u *UsersController) GetUsersWithGLAccounts(c echo.Context) error {
	users, err := u.UsersService.GetUsersWithGLAccounts()
	if err != nil {
		log.Printf("Error getting users with GLAccounts: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}

func (u *UsersController) ChangePassword(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	request := new(dtos.ChangePasswordRequestDTO)
	if err := c.Bind(request); err != nil {
		log.Printf("Error binding change password request: %v", err.Error())
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.ChangePassword(uint(id), request); err != nil {
		log.Printf("Error changing password: %v", err.Error())
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, "Password changed successfully")
}

func (u *UsersController) ForgotPassword(c echo.Context) error {
	request := new(dtos.PasswordResetRequestDTO)
	if err := c.Bind(request); err != nil {
		log.Printf("Error binding forget password request: %v", err.Error())
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.ForgotPassword(request); err != nil {
		log.Printf("Error forgot password: %v", err.Error())
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset link sent successfully"})
}

func (u *UsersController) ValidatePasswordResetToken(c echo.Context) error {
	token := new(dtos.PasswordResetTokenDTO)
	if err := c.Bind(token); err != nil {
		log.Printf("Error binding password reset token: %v", err.Error())
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.ValidatePasswordResetToken(*token); err != nil {
		log.Printf("Error validating password reset token: %v", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]any{"message": err.Error(), "valid": false})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Password reset token is valid", "valid": true})
}

func (u *UsersController) ResetPassword(c echo.Context) error {
	request := new(dtos.PasswordResetChangeRequestDTO)
	if err := c.Bind(request); err != nil {
		log.Printf("Error binding forget password change request: %v", err.Error())
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.ResetPassword(request); err != nil {
		log.Printf("Error forgot password change: %v", err.Error())
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Password changed successfully"})
}
