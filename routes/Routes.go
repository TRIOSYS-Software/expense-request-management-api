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

	expenseApprovalsRepo := repositories.NewExpenseApprovalsRepo(db)
	expenseApprovalsService := services.NewExpenseApprovalsService(expenseApprovalsRepo)
	expenseApprovalsController := controllers.NewExpenseApprovalsController(expenseApprovalsService)
	expenseApprovalsRoutes(apiV1, expenseApprovalsController)
}

func usersRoutes(e *echo.Group, controllers *controllers.UsersController) {
	e.GET("/users", controllers.GetUsers, middlewares.IsAuthenticated)
	e.POST("/users", controllers.CreateUser, middlewares.IsAdmin)
	e.POST("/login", controllers.LoginUser)
	e.GET("/users/:id", controllers.GetUserByID, middlewares.IsAuthenticated)
	e.GET("/users/role/:role_id", controllers.GetUsersByRole, middlewares.IsAuthenticated)
	e.PUT("/users/:id", controllers.UpdateUser, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/users/:id", controllers.DeleteUser, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.POST("/verify", controllers.VerifyUser, middlewares.IsAuthenticated)
}

func departmentsRoutes(e *echo.Group, controllers *controllers.DepartmentsController) {
	e.GET("/departments", controllers.GetDepartments, middlewares.IsAuthenticated)
	e.POST("/departments", controllers.CreateDepartment, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/departments/:id", controllers.GetDepartmentByID, middlewares.IsAuthenticated)
	e.PUT("/departments/:id", controllers.UpdateDepartment, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/departments/:id", controllers.DeleteDepartment, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func rolesRoutes(e *echo.Group, controllers *controllers.RolesController) {
	e.GET("/roles", controllers.GetRoles, middlewares.IsAuthenticated)
	e.POST("/roles", controllers.CreateRole, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/roles/:id", controllers.GetRoleByID, middlewares.IsAuthenticated)
	e.PUT("/roles/:id", controllers.UpdateRole, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/roles/:id", controllers.DeleteRole, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func approvalPoliciesRoutes(e *echo.Group, controllers *controllers.ApprovalPoliciesController) {
	e.GET("/approval-policies", controllers.GetApprovalPolicies, middlewares.IsAuthenticated)
	e.POST("/approval-policies", controllers.CreateApprovalPolicy, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/approval-policies/:id", controllers.GetApprovalPolicyByID, middlewares.IsAuthenticated)
	e.PUT("/approval-policies/:id", controllers.UpdateApprovalPolicy, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/approval-policies/:id", controllers.DeleteApprovalPolicy, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func expenseCategoriesRoutes(e *echo.Group, controllers *controllers.ExpenseCategoriesController) {
	e.GET("/expense-categories", controllers.GetExpenseCategories, middlewares.IsAuthenticated)
	e.POST("/expense-categories", controllers.CreateExpenseCategory, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/expense-categories/:id", controllers.GetExpenseCategoryByID, middlewares.IsAuthenticated)
	e.PUT("/expense-categories/:id", controllers.UpdateExpenseCategory, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/expense-categories/:id", controllers.DeleteExpenseCategory, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func expenseRequestsRoutes(e *echo.Group, controllers *controllers.ExpenseRequestsController) {
	e.GET("/expense-requests", controllers.GetExpenseRequests, middlewares.IsAuthenticated)
	e.POST("/expense-requests", controllers.CreateExpenseRequest, middlewares.IsAuthenticated)
	e.GET("/expense-requests/:id", controllers.GetExpenseRequestByID, middlewares.IsAuthenticated)
	e.GET("/expense-requests/user/:id", controllers.GetExpenseRequestsByUserID, middlewares.IsAuthenticated)
	e.GET("/expense-requests/approver/:id", controllers.GetExpenseRequestByApproverID, middlewares.IsAuthenticated)
}

func expenseApprovalsRoutes(e *echo.Group, controllers *controllers.ExpenseApprovalsController) {
	e.GET("/expense-approvals", controllers.GetExpenseApprovals, middlewares.IsAuthenticated)
	e.GET("/expense-approvals/approver/:approver_id", controllers.GetExpenseApprovalsByApproverID, middlewares.IsAuthenticated)
	e.PUT("/expense-approvals/:id", controllers.UpdateExpenseApproval, middlewares.IsAuthenticated)
}
