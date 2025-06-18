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

// GetUsers go doc
// @Summary Get all users
// @Description Get all users
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} []models.Users
// @Failure 500 {object} string
// @Security JWT Token
// @Router /users [get]
func (u *UsersController) GetUsers(c echo.Context) error {
	users, err := u.UsersService.GetUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
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

// LoginUser go doc
// @Summary Login user
// @Description Login user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dtos.LoginRequestDTO true "User"
// @Success 200 {object} dtos.LoginResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} string
// @Router /login [post]
func (u *UsersController) LoginUser(c echo.Context) error {
	resquest := new(dtos.LoginRequestDTO)
	if err := c.Bind(resquest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	user := models.Users{
		Email:    resquest.Email,
		Password: resquest.Password,
	}
	data, err := u.UsersService.LoginUser(&user)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}
	return c.JSON(http.StatusOK, data)
}

// CreateUser go doc
// @Summary Create user
// @Description Create user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dtos.UserRequestDTO true "User"
// @Success 200 {object} models.Users
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users [post]
func (u *UsersController) CreateUser(c echo.Context) error {
	request := new(dtos.UserRequestDTO)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	user := models.Users{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		RoleID:   request.Role,
	}
	if err := u.UsersService.CreateUser(&user); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, user)
}

// GetUserByID go doc
// @Summary Get user by id
// @Description Get user by id
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.Users
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/{id} [get]
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

// UpdateUser go doc
// @Summary Update user
// @Description Update user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dtos.UserRequestDTO true "User"
// @Success 200 {object} models.Users
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/{id} [put]
func (u *UsersController) UpdateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}
	user := new(models.Users)
	user.ID = uint(id)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := u.UsersService.UpdateUser(user); err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

// DeleteUser go doc
// @Summary Delete user
// @Description Delete user
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/{id} [delete]
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

// VerifyUser go doc
// @Summary Verify user
// @Description Verify user
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} models.Users
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /verify [post]
func (u *UsersController) VerifyUser(c echo.Context) error {
	user_id := c.Get("user_id")
	user, err := u.UsersService.GetUserByID(uint(user_id.(float64)))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid User!")
	}
	return c.JSON(http.StatusOK, user)
}

// SetPaymentMethodsToUser go doc
// @Summary Set payment methods to user
// @Description Set payment methods to user
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security JWT Token
// @Router /users/set-payment-methods [post]
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

// GetUsersWithPaymentMethods go doc
// @Summary Get users with payment methods
// @Description Get users with payment methods
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} []models.Users
// @Failure 500 {object} string
// @Security JWT Token
// @Router /users/payment-methods [get]
func (u *UsersController) GetUsersWithPaymentMethods(c echo.Context) error {
	users, err := u.UsersService.GetUsersWithPaymentMethods()
	if err != nil {
		log.Printf("Error getting users with payment methods: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}

// GetPaymentMethodsByUserID go doc
// @Summary Get payment methods by user id
// @Description Get payment methods by user id
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} []models.PaymentMethod
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/{id}/payment-methods [get]
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

// SetGLAccountsToUser go doc
// @Summary Set GLAccounts to user
// @Description Set GLAccounts to user
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security JWT Token
// @Router /users/set-gl-accounts [post]
func (u *UsersController) SetGLAccountsToUser(c echo.Context) error {
	userGLAccountDTO := new(dtos.UserGLAccountDTO)
	if err := c.Bind(userGLAccountDTO); err != nil {
		log.Printf("Error binding GLAccounts to user: %v", err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := u.UsersService.SetGLAccountsToUser(userGLAccountDTO); err != nil {
		log.Printf("Error setting GLAccounts to user: %v", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "GLAccounts set successfully")
}

// GetGLAccountsByUserID go doc
// @Summary Get GLAccounts by user id
// @Description Get GLAccounts by user id
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} []models.GLAcc
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/{id}/gl-accounts [get]
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

// GetUsersWithGLAccounts go doc
// @Summary Get users with GLAccounts
// @Description Get users with GLAccounts
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} []models.Users
// @Failure 500 {object} string
// @Security JWT Token
// @Router /users/gl-accounts [get]
func (u *UsersController) GetUsersWithGLAccounts(c echo.Context) error {
	users, err := u.UsersService.GetUsersWithGLAccounts()
	if err != nil {
		log.Printf("Error getting users with GLAccounts: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}

// ChangePassword go doc
// @Summary Change password
// @Description Change password
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body dtos.ChangePasswordRequestDTO true "Change password request"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/{id}/change-password [put]
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

// ForgotPassword go doc
// @Summary Forgot password
// @Description Forgot password
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dtos.PasswordResetRequestDTO true "Password reset request"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /forgot-password [post]
func (u *UsersController) ForgotPassword(c echo.Context) error {
	request := new(dtos.PasswordResetRequestDTO)
	if err := c.Bind(request); err != nil {
		log.Printf("Error binding forget password request: %v", err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := u.UsersService.ForgotPassword(request); err != nil {
		log.Printf("Error forgot password: %v", err.Error())
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset link sent successfully"})
}

// ValidatePasswordResetToken go doc
// @Summary Validate password reset token
// @Description Validate password reset token
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dtos.PasswordResetTokenDTO true "Password reset token"
// @Success 200 {object} map[string]any
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /validate-password-reset-token [post]
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

// ResetPassword go doc
// @Summary Reset password
// @Description Reset password
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dtos.PasswordResetChangeRequestDTO true "Password reset change request"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /reset-password [post]
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

// SetProjectsToUser go doc
// @Summary Set projects to user
// @Description Set projects to user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dtos.UserProjectDTO true "User project"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/set-projects [post]
func (u *UsersController) SetProjectsToUser(c echo.Context) error {
	userProjectDTO := new(dtos.UserProjectDTO)
	if err := c.Bind(userProjectDTO); err != nil {
		log.Printf("Error binding projects to user: %v", err)
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := u.UsersService.SetProjectsToUser(userProjectDTO); err != nil {
		log.Printf("Error setting projects to user: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, "Projects set successfully")
}

// GetProjectsByUserID go doc
// @Summary Get projects by user id
// @Description Get projects by user id
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User id"
// @Success 200 {object} []models.Project
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/{id}/projects [get]
func (u *UsersController) GetProjectsByUserID(ctx echo.Context) error {
	userID := ctx.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid user id")
	}
	projects, err := u.UsersService.GetUserProjects(uint(id))
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusNotFound, err)
	}
	return ctx.JSON(http.StatusOK, projects)
}

// GetUsersWithProjects go doc
// @Summary Get users with projects
// @Description Get users with projects
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} []models.Users
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Security JWT Token
// @Router /users/projects [get]
func (u *UsersController) GetUsersWithProjects(c echo.Context) error {
	users, err := u.UsersService.GetUsersWithProjects()
	if err != nil {
		log.Printf("Error getting users with projects: %v", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}
