package routes

import (
	"shwetaik-expense-management-api/controllers"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/services"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func InitialRoute(e *echo.Echo, db *gorm.DB) {
	apiV1 := e.Group("/api/v1")

	usersRepo := repositories.NewUsersRepo(db)
	usersService := services.NewUsersService(usersRepo)
	usersController := controllers.NewUsersController(usersService)
	usersRoutes(apiV1, usersController)

	departmentsRepo := repositories.NewDepartmentsRepo(db)
	departmentsService := services.NewDepartmentsService(departmentsRepo)
	departmentsController := controllers.NewDepartmentsController(departmentsService)
	departmentsRoutes(apiV1, departmentsController)

	rolesRepo := repositories.NewRolesRepo(db)
	rolesService := services.NewRolesService(rolesRepo)
	rolesController := controllers.NewRolesController(rolesService)
	rolesRoutes(apiV1, rolesController)
}

func usersRoutes(e *echo.Group, controllers *controllers.UsersController) {
	e.GET("/users", controllers.GetUsers)
	e.POST("/users", controllers.CreateUser)
	e.GET("/users/:id", controllers.GetUserByID)
	e.PUT("/users/:id", controllers.UpdateUser)
	e.DELETE("/users/:id", controllers.DeleteUser)
}

func departmentsRoutes(e *echo.Group, controllers *controllers.DepartmentsController) {
	e.GET("/departments", controllers.GetDepartments)
	e.POST("/departments", controllers.CreateDepartment)
	e.GET("/departments/:id", controllers.GetDepartmentByID)
}

func rolesRoutes(e *echo.Group, controllers *controllers.RolesController) {
	e.GET("/roles", controllers.GetRoles)
	e.POST("/roles", controllers.CreateRole)
	e.GET("/roles/:id", controllers.GetRoleByID)
}
