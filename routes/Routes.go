package routes

import (
	"shwetaik-expense-management-api/controllers"
	"shwetaik-expense-management-api/middlewares"
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

	approvalPoliciesRepo := repositories.NewApprovalPoliciesRepo(db)
	approvalPoliciesService := services.NewApprovalPoliciesService(approvalPoliciesRepo)
	approvalPoliciesController := controllers.NewApprovalPoliciesController(approvalPoliciesService)
	approvalPoliciesRoutes(apiV1, approvalPoliciesController)

	expenseCategoriesRepo := repositories.NewExpenseCategoriesRepo(db)
	expenseCategoriesService := services.NewExpenseCategoriesService(expenseCategoriesRepo)
	expenseCategoriesController := controllers.NewExpenseCategoriesController(expenseCategoriesService)
	expenseCategoriesRoutes(apiV1, expenseCategoriesController)

	expenseRequestsRepo := repositories.NewExpenseRequestsRepo(db)
	expenseRequestsService := services.NewExpenseRequestsService(expenseRequestsRepo)
	expenseRequestsController := controllers.NewExpenseRequestsController(expenseRequestsService)
	expenseRequestsRoutes(apiV1, expenseRequestsController)
}

func usersRoutes(e *echo.Group, controllers *controllers.UsersController) {
	e.GET("/users", controllers.GetUsers, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.POST("/users", controllers.CreateUser)
	e.POST("/login", controllers.LoginUser)
	e.GET("/users/:id", controllers.GetUserByID)
	e.PUT("/users/:id", controllers.UpdateUser)
	e.DELETE("/users/:id", controllers.DeleteUser)
	e.POST("/verify", controllers.VerifyUser, middlewares.IsAuthenticated)
}

func departmentsRoutes(e *echo.Group, controllers *controllers.DepartmentsController) {
	e.GET("/departments", controllers.GetDepartments)
	e.POST("/departments", controllers.CreateDepartment)
	e.GET("/departments/:id", controllers.GetDepartmentByID)
	e.PUT("/departments/:id", controllers.UpdateDepartment)
	e.DELETE("/departments/:id", controllers.DeleteDepartment)
}

func rolesRoutes(e *echo.Group, controllers *controllers.RolesController) {
	e.GET("/roles", controllers.GetRoles)
	e.POST("/roles", controllers.CreateRole)
	e.GET("/roles/:id", controllers.GetRoleByID)
	e.PUT("/roles/:id", controllers.UpdateRole)
	e.DELETE("/roles/:id", controllers.DeleteRole)
}

func approvalPoliciesRoutes(e *echo.Group, controllers *controllers.ApprovalPoliciesController) {
	e.GET("/approval-policies", controllers.GetApprovalPolicies)
	e.POST("/approval-policies", controllers.CreateApprovalPolicy)
	e.GET("/approval-policies/:id", controllers.GetApprovalPolicyByID)
	e.PUT("/approval-policies/:id", controllers.UpdateApprovalPolicy)
	e.DELETE("/approval-policies/:id", controllers.DeleteApprovalPolicy)
}

func expenseCategoriesRoutes(e *echo.Group, controllers *controllers.ExpenseCategoriesController) {
	e.GET("/expense-categories", controllers.GetExpenseCategories)
	e.POST("/expense-categories", controllers.CreateExpenseCategory)
	e.GET("/expense-categories/:id", controllers.GetExpenseCategoryByID)
	e.PUT("/expense-categories/:id", controllers.UpdateExpenseCategory)
	e.DELETE("/expense-categories/:id", controllers.DeleteExpenseCategory)
}

func expenseRequestsRoutes(e *echo.Group, controllers *controllers.ExpenseRequestsController) {
	e.GET("/expense-requests", controllers.GetExpenseRequests)
	e.POST("/expense-requests", controllers.CreateExpenseRequest)
	e.GET("/expense-requests/:id", controllers.GetExpenseRequestByID)
}
