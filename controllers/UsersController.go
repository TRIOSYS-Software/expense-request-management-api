package controllers

import (
	"net/http"
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

func (u *UsersController) LoginUser(c echo.Context) error {
	user := new(models.Users)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	data, err := u.UsersService.LoginUser(user)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
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
	user := new(models.Users)
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
