package routes

import (
	"strings"
	"time"
	"utara_backend/handlers"
	"utara_backend/middleware"
	"utara_backend/models"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	// origin := "https://utara-app.web.app"
	// origin := "http://localhost:53505" // Change this to your frontend URL
	// CORS config
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "https://utara-app.web.app")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/signup", handlers.Signup)
		auth.POST("/verify-signup-otp", handlers.VerifySignupOTP)
		auth.POST("/login", handlers.Login)
		auth.POST("/user-login", handlers.UserLogin)
		auth.POST("/verify-otp", handlers.VerifyOTP)
		auth.POST("/forgot-password", handlers.ForgotPassword)
		auth.POST("/reset-password", handlers.ResetPassword)
	}

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/createUser", handlers.CreateUser)

		protected.GET("/profile", handlers.GetProfile)
		protected.GET("/users", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.GetAllUsers)
		protected.POST("/assign-module", middleware.RequireRole(models.RoleSuperAdmin), handlers.AssignModulesHandler)
		protected.POST("/assign-usertype", middleware.RequireRole(models.RoleSuperAdmin), handlers.AssignUserType)

		// Users routes
		users := protected.Group("/user")
		{
			users.PUT("/update-user/:id", middleware.RequireRole(models.RoleSuperAdmin), handlers.UpdateUsers)
			users.DELETE("/delete-user/:id", middleware.RequireRole(models.RoleSuperAdmin), handlers.DeleteUser)
		}

		// Room routes
		rooms := protected.Group("/rooms")
		{
			rooms.POST("/", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.CreateRoom)
			rooms.GET("/", handlers.GetRooms)
			rooms.GET("/stats", handlers.GetRoomStats)
			rooms.GET("/:id", handlers.GetRoom)
			rooms.PUT("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.UpdateRoom)
			rooms.POST("/upload-rooms", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.CreateMultipleRooms)
			rooms.DELETE("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.DeleteRoom)
			rooms.GET("/buildings", handlers.GetBuildings)
			rooms.GET("/floors", handlers.GetFloors)

			//Room Category
			rooms.POST("/create-room-category", middleware.RequireRole(models.RoleSuperAdmin), handlers.CreateRoomCategory)
			rooms.GET("/get-room-categories", middleware.RequireRole(models.RoleSuperAdmin), handlers.GetRoomCategories)
			rooms.PUT("/update-room-category/:id", middleware.RequireRole(models.RoleSuperAdmin), handlers.UpdateRoomCategory)
			rooms.DELETE("/delete-room-category/:id", middleware.RequireRole(models.RoleSuperAdmin), handlers.DeleteRoomCategory)
		}

		// Room request routes
		requests := protected.Group("/room-requests")
		{
			requests.POST("/", handlers.CreateRoomRequest)
			requests.GET("/", handlers.GetRoomRequests)
			requests.PUT("/:id/process", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.ProcessRoomRequest)
		}

		// Room assignment routes
		assignments := protected.Group("/room-assignments")
		{
			assignments.POST("/", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.AssignRoom)
			assignments.PUT("/:id/check-in", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.CheckInRoom)
			assignments.PUT("/:id/check-out", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.CheckOutRoom)
		}

		// Food pass routes
		foodPasses := protected.Group("/food-passes")
		{
			foodPasses.POST("/generate", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.GenerateFoodPasses)
			foodPasses.GET("/user/:user_id", handlers.GetUserFoodPasses)
			foodPasses.POST("/scan", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.ScanFoodPass)
			foodPasses.PUT("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.UpdateFoodPass)

			//Food pass category
			foodPasses.POST("/food-pass-category", middleware.RequireRole(models.RoleSuperAdmin), handlers.CreateFoodPassCategory)
			foodPasses.GET("/get-pass-categories", middleware.RequireRole(models.RoleSuperAdmin), handlers.GetFoodPassCategories)
			foodPasses.PUT("/update-pass-category/:id", middleware.RequireRole(models.RoleSuperAdmin), handlers.UpdateFoodPassCategory)
			foodPasses.DELETE("/delete-pass-category/:id", middleware.RequireRole(models.RoleSuperAdmin), handlers.DeleteFoodPassCategory)
		}
	}
}
