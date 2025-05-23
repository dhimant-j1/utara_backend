package routes

import (
	"time"

	"utara_backend/handlers"
	"utara_backend/middleware"
	"utara_backend/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	// CORS config
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:50855"},
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
		auth.POST("/login", handlers.Login)
	}

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/createUser", handlers.CreateUser)

		protected.GET("/profile", handlers.GetProfile)

		// Room routes
		rooms := protected.Group("/rooms")
		{
			rooms.POST("/", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.CreateRoom)
			rooms.GET("/", handlers.GetRooms)
			rooms.GET("/stats", handlers.GetRoomStats)
			rooms.GET("/:id", handlers.GetRoom)
			rooms.PUT("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleStaff), handlers.UpdateRoom)
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
		}
	}
}
