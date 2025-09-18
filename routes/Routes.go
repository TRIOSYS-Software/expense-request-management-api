package routes

import (
	"log"
	"shwetaik-expense-management-api/controllers"
	"shwetaik-expense-management-api/middlewares"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/services"

	firebase "firebase.google.com/go/v4"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func InitialRoute(e *echo.Echo, db *gorm.DB, firebaseApp *firebase.App) {
	apiV1 := e.Group("/api/v1")

	initUsersRoutes(apiV1, db)
	initDepartmentsRoutes(apiV1, db)
	initRolesRoutes(apiV1, db)
	initApprovalPoliciesRoutes(apiV1, db)
	initExpenseCategoriesRoutes(apiV1, db)
	initExpenseRequestsRoutes(apiV1, db, firebaseApp)
	initExpenseApprovalsRoutes(apiV1, db, firebaseApp)
	initNotificationRoutes(apiV1, db, firebaseApp)
	initPaymentMethodsRoutes(apiV1, db)
	initProjectsRoutes(apiV1, db)
	initGLAccRoutes(apiV1, db)

	initDeviceTokenRoutes(apiV1, db)
	initWebsocketRoutes(e)
}

func initUsersRoutes(e *echo.Group, db *gorm.DB) {
	usersRepo := repositories.NewUsersRepo(db)
	usersService := services.NewUsersService(usersRepo)
	usersController := controllers.NewUsersController(usersService)
	e.GET("/users", usersController.GetUsers, middlewares.IsAuthenticated)
	e.POST("/users", usersController.CreateUser, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.POST("/login", usersController.LoginUser)
	e.GET("/users/:id", usersController.GetUserByID, middlewares.IsAuthenticated)
	e.GET("/users/role/:role_id", usersController.GetUsersByRole, middlewares.IsAuthenticated)
	e.PUT("/users/:id", usersController.UpdateUser, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/users/:id", usersController.DeleteUser, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.POST("/verify", usersController.VerifyUser, middlewares.IsAuthenticated)
	e.POST("/users/set-payment-methods", usersController.SetPaymentMethodsToUser, middlewares.IsAuthenticated)
	e.GET("/users/payment-methods", usersController.GetUsersWithPaymentMethods, middlewares.IsAuthenticated)
	e.GET("/users/:id/payment-methods", usersController.GetPaymentMethodsByUserID, middlewares.IsAuthenticated)
	e.POST("/users/set-gl-accounts", usersController.SetGLAccountsToUser, middlewares.IsAuthenticated)
	e.GET("/users/gl-accounts", usersController.GetUsersWithGLAccounts, middlewares.IsAuthenticated)
	e.GET("/users/:id/gl-accounts", usersController.GetGLAccountsByUserID, middlewares.IsAuthenticated)
	e.PUT("/users/:id/change-password", usersController.ChangePassword, middlewares.IsAuthenticated)
	e.POST("/forgot-password", usersController.ForgotPassword)
	e.POST("/validate-password-reset-token", usersController.ValidatePasswordResetToken)
	e.POST("/reset-password", usersController.ResetPassword)
	e.POST("/users/set-projects", usersController.SetProjectsToUser, middlewares.IsAuthenticated)
	e.GET("/users/projects", usersController.GetUsersWithProjects, middlewares.IsAuthenticated)
	e.GET("/users/:id/projects", usersController.GetProjectsByUserID, middlewares.IsAuthenticated)
}

func initDepartmentsRoutes(e *echo.Group, db *gorm.DB) {
	departmentsRepo := repositories.NewDepartmentsRepo(db)
	departmentsService := services.NewDepartmentsService(departmentsRepo)
	departmentsController := controllers.NewDepartmentsController(departmentsService)
	e.GET("/departments", departmentsController.GetDepartments, middlewares.IsAuthenticated)
	e.POST("/departments", departmentsController.CreateDepartment, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/departments/:id", departmentsController.GetDepartmentByID, middlewares.IsAuthenticated)
	e.PUT("/departments/:id", departmentsController.UpdateDepartment, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/departments/:id", departmentsController.DeleteDepartment, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func initRolesRoutes(e *echo.Group, db *gorm.DB) {
	rolesRepo := repositories.NewRolesRepo(db)
	rolesService := services.NewRolesService(rolesRepo)
	rolesController := controllers.NewRolesController(rolesService)
	e.GET("/roles", rolesController.GetRoles, middlewares.IsAuthenticated)
	e.POST("/roles", rolesController.CreateRole, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/roles/:id", rolesController.GetRoleByID, middlewares.IsAuthenticated)
	e.PUT("/roles/:id", rolesController.UpdateRole, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/roles/:id", rolesController.DeleteRole, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func initApprovalPoliciesRoutes(e *echo.Group, db *gorm.DB) {
	approvalPoliciesRepo := repositories.NewApprovalPoliciesRepo(db)
	approvalPoliciesService := services.NewApprovalPoliciesService(approvalPoliciesRepo)
	approvalPoliciesController := controllers.NewApprovalPoliciesController(approvalPoliciesService)
	e.GET("/approval-policies", approvalPoliciesController.GetApprovalPolicies, middlewares.IsAuthenticated)
	e.POST("/approval-policies", approvalPoliciesController.CreateApprovalPolicy, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/approval-policies/:id", approvalPoliciesController.GetApprovalPolicyByID, middlewares.IsAuthenticated)
	e.PUT("/approval-policies/:id", approvalPoliciesController.UpdateApprovalPolicy, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/approval-policies/:id", approvalPoliciesController.DeleteApprovalPolicy, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func initExpenseCategoriesRoutes(e *echo.Group, db *gorm.DB) {
	expenseCategoriesRepo := repositories.NewExpenseCategoriesRepo(db)
	expenseCategoriesService := services.NewExpenseCategoriesService(expenseCategoriesRepo)
	expenseCategoriesController := controllers.NewExpenseCategoriesController(expenseCategoriesService)
	e.GET("/expense-categories", expenseCategoriesController.GetExpenseCategories, middlewares.IsAuthenticated)
	e.POST("/expense-categories", expenseCategoriesController.CreateExpenseCategory, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/expense-categories/:id", expenseCategoriesController.GetExpenseCategoryByID, middlewares.IsAuthenticated)
	e.PUT("/expense-categories/:id", expenseCategoriesController.UpdateExpenseCategory, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.DELETE("/expense-categories/:id", expenseCategoriesController.DeleteExpenseCategory, middlewares.IsAuthenticated, middlewares.IsAdmin)
}

func initExpenseRequestsRoutes(e *echo.Group, db *gorm.DB, firebaseApp *firebase.App) {
	expenseRequestsRepo := repositories.NewExpenseRequestsRepo(db, firebaseApp)
	expenseRequestsService := services.NewExpenseRequestsService(expenseRequestsRepo)
	expenseRequestsController := controllers.NewExpenseRequestsController(expenseRequestsService)
	e.GET("/expense-requests", expenseRequestsController.GetExpenseRequests, middlewares.IsAuthenticated)
	e.POST("/expense-requests", expenseRequestsController.CreateExpenseRequest, middlewares.IsAuthenticated)
	e.PUT("/expense-requests/:id", expenseRequestsController.UpdateExpenseRequest, middlewares.IsAuthenticated)
	e.GET("/expense-requests/:id", expenseRequestsController.GetExpenseRequestByID, middlewares.IsAuthenticated)
	e.GET("/expense-requests/user/:id", expenseRequestsController.GetExpenseRequestsByUserID, middlewares.IsAuthenticated)
	e.GET("/expense-requests/approver/:id", expenseRequestsController.GetExpenseRequestByApproverID, middlewares.IsAuthenticated)
	e.GET("/expense-requests/summary", expenseRequestsController.GetExpenseRequestsSummary, middlewares.IsAuthenticated)
	e.POST("/expense-requests/:id/send-to-sqlacc", expenseRequestsController.SendExpenseRequestToSQLACC, middlewares.IsAuthenticated)
	e.DELETE("/expense-requests/:id", expenseRequestsController.DeleteExpenseRequest, middlewares.IsAuthenticated)
	e.GET("/expense-requests/attachment/:filename", expenseRequestsController.ServeExpenseRequestAttachment)
}

func initExpenseApprovalsRoutes(e *echo.Group, db *gorm.DB, firebaseApp *firebase.App) {
	expenseApprovalsRepo := repositories.NewExpenseApprovalsRepo(db, firebaseApp)
	expenseApprovalsService := services.NewExpenseApprovalsService(expenseApprovalsRepo)
	expenseApprovalsController := controllers.NewExpenseApprovalsController(expenseApprovalsService)
	e.GET("/expense-approvals", expenseApprovalsController.GetExpenseApprovals, middlewares.IsAuthenticated)
	e.GET("/expense-approvals/approver/:approver_id", expenseApprovalsController.GetExpenseApprovalsByApproverID, middlewares.IsAuthenticated)
	e.PUT("/expense-approvals/:id", expenseApprovalsController.UpdateExpenseApproval, middlewares.IsAuthenticated)
}

func initNotificationRoutes(e *echo.Group, db *gorm.DB, firebaseApp *firebase.App) {
	notificationRepo := repositories.NewNotificationRepo(db, firebaseApp)
	notificationService := services.NewNotificationService(notificationRepo)
	notificationController := controllers.NewNotificationController(notificationService)

	e.GET("/notifications/user/:userID", notificationController.GetNotificationsByUserID, middlewares.IsAuthenticated)
	e.PUT("/notifications/:id/read", notificationController.MarkNotificationAsRead, middlewares.IsAuthenticated)
	e.PUT("/notifications/user/:userID/mark-all-read", notificationController.MarkAllNotificationsAsRead, middlewares.IsAuthenticated)
	e.DELETE("/notifications/:id", notificationController.DeleteNotification, middlewares.IsAuthenticated)
	e.DELETE("/notifications/user/:userID/clear", notificationController.ClearAllNotifications, middlewares.IsAuthenticated)
}

func initPaymentMethodsRoutes(e *echo.Group, db *gorm.DB) {
	paymentMethodRepo := repositories.NewPaymentMethodRepo(db)
	paymentMethodService := services.NewPaymentMethodService(paymentMethodRepo)
	paymentMethodController := controllers.NewPaymentMethodController(paymentMethodService)
	e.POST("/payment-methods/sync", paymentMethodController.SyncPaymentMethods, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/payment-methods", paymentMethodController.GetPaymentMethods, middlewares.IsAuthenticated)

	go func() {
		err := paymentMethodService.SyncPaymentMethods()
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Println("Payment methods synced successfully")
		}
	}()
}

func initProjectsRoutes(e *echo.Group, db *gorm.DB) {
	projectRepo := repositories.NewProjectRepo(db)
	projectService := services.NewProjectService(projectRepo)
	projectController := controllers.NewProjectController(projectService)
	e.POST("/projects/sync", projectController.SyncProjects, middlewares.IsAuthenticated, middlewares.IsAdmin)
	e.GET("/projects", projectController.GetProjects, middlewares.IsAuthenticated)

	go func() {
		err := projectService.SyncProjects()
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Println("Projects synced successfully")
		}
	}()
}

func initGLAccRoutes(e *echo.Group, db *gorm.DB) {
	glAccRepo := repositories.NewGLAccRepo(db)
	glAccService := services.NewGLAccService(glAccRepo)
	glAccController := controllers.NewGLAccController(glAccService)
	e.GET("/gl-acc", glAccController.GetGLAcc, middlewares.IsAuthenticated)
	e.POST("/gl-acc/sync", glAccController.SyncGLAcc, middlewares.IsAuthenticated, middlewares.IsAdmin)

	go func() {
		err := glAccService.SyncGLAcc()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("GLAcc synced successfully")
		}
	}()
}

func initWebsocketRoutes(e *echo.Echo) {
	e.GET("/ws/:userID", controllers.HandleWebSocket)
}

func initDeviceTokenRoutes(e *echo.Group, db *gorm.DB) {
	deviceTokenRepo := repositories.NewDeviceTokenRepo(db)
	deviceTokenService := services.NewDeviceTokenService(deviceTokenRepo)
	deviceTokenController := controllers.NewDeviceTokenController(deviceTokenService)
	

	e.GET("/users/:id/device-tokens", deviceTokenController.GetTokensByUserID, middlewares.IsAuthenticated)
	e.POST("/users/:id/device-tokens", deviceTokenController.CreateTokenByUserID)
}