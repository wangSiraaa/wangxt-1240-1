package main

import (
	"log"

	"grain-management/config"
	"grain-management/controllers"
	"grain-management/database"
	"grain-management/middleware"
	"grain-management/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	database.DB.AutoMigrate(
		&models.User{},
		&models.Granary{},
		&models.Sensor{},
		&models.GrainConditionRecord{},
		&models.SensorReading{},
		&models.FumigationPlan{},
		&models.FumigationExecution{},
		&models.UnsealRecord{},
		&models.GasDetectionRecord{},
		&models.GrainTurnoverSuggestion{},
		&models.OperationLog{},
	)

	authCtrl := controllers.NewAuthController(cfg)
	granaryCtrl := controllers.NewGranaryController()
	grainCondCtrl := controllers.NewGrainConditionController()
	fumigationCtrl := controllers.NewFumigationController()
	unsealCtrl := controllers.NewUnsealController()
	suggestionCtrl := controllers.NewTurnoverSuggestionController()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authCtrl.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg), authCtrl.GetCurrentUser)
		}

		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(cfg))
		{
			authenticated.GET("/dashboard/stats", suggestionCtrl.GetDashboardStats)

			granaries := authenticated.Group("/granaries")
			{
				granaries.GET("", granaryCtrl.List)
				granaries.GET("/keepers", granaryCtrl.ListKeepers)
				granaries.GET("/:id", granaryCtrl.Get)
				granaries.POST("", middleware.RoleMiddleware(models.RoleAdmin, models.RoleKeeper), granaryCtrl.Create)
				granaries.PUT("/:id", middleware.RoleMiddleware(models.RoleAdmin, models.RoleKeeper), granaryCtrl.Update)
				granaries.DELETE("/:id", middleware.RoleMiddleware(models.RoleAdmin), granaryCtrl.Delete)
				granaries.GET("/:id/sensors", granaryCtrl.GetSensors)
				granaries.POST("/:id/sensors", middleware.RoleMiddleware(models.RoleAdmin, models.RoleKeeper), granaryCtrl.AddSensor)
				granaries.POST("/:id/sensors/:sensorId/readings", granaryCtrl.AddSensorReading)
				granaries.GET("/:id/readings", grainCondCtrl.GetSensorReadings)
			}

			grainCond := authenticated.Group("/grain-conditions")
			{
				grainCond.GET("", grainCondCtrl.List)
				grainCond.GET("/:id", grainCondCtrl.Get)
				grainCond.POST("", middleware.RoleMiddleware(models.RoleKeeper, models.RoleAdmin), grainCondCtrl.Create)
			}

			fumigation := authenticated.Group("/fumigation")
			{
				fumigation.GET("/plans", fumigationCtrl.List)
				fumigation.GET("/plans/:id", fumigationCtrl.Get)
				fumigation.POST("/plans", middleware.RoleMiddleware(models.RoleKeeper, models.RoleAdmin), fumigationCtrl.Create)
				fumigation.POST("/plans/:id/submit", middleware.RoleMiddleware(models.RoleKeeper, models.RoleAdmin), fumigationCtrl.SubmitForApproval)
				fumigation.POST("/plans/:id/approve", middleware.RoleMiddleware(models.RoleSafetyOfficer, models.RoleAdmin), fumigationCtrl.Approve)
				fumigation.POST("/plans/:id/clear-people", middleware.RoleMiddleware(models.RoleKeeper, models.RoleAdmin, models.RoleDutyOfficer), fumigationCtrl.MarkPeopleCleared)
				fumigation.POST("/plans/:id/start", middleware.RoleMiddleware(models.RoleKeeper, models.RoleAdmin), fumigationCtrl.StartExecution)
				fumigation.POST("/plans/:id/complete", middleware.RoleMiddleware(models.RoleKeeper, models.RoleAdmin), fumigationCtrl.CompleteExecution)
				fumigation.GET("/executions", fumigationCtrl.ListExecutions)
			}

			unseal := authenticated.Group("/unseal")
			{
				unseal.GET("", unsealCtrl.List)
				unseal.GET("/:id", unsealCtrl.Get)
				unseal.POST("", middleware.RoleMiddleware(models.RoleDutyOfficer, models.RoleKeeper, models.RoleAdmin), unsealCtrl.Create)
				unseal.POST("/:id/complete", middleware.RoleMiddleware(models.RoleDutyOfficer, models.RoleAdmin), unsealCtrl.CompleteUnseal)
				unseal.GET("/gas-detections", unsealCtrl.ListGasDetections)
				unseal.POST("/:id/gas-detections", middleware.RoleMiddleware(models.RoleDutyOfficer, models.RoleAdmin), unsealCtrl.AddGasDetection)
			}

			suggestions := authenticated.Group("/turnover-suggestions")
			{
				suggestions.GET("", suggestionCtrl.List)
				suggestions.GET("/:id", suggestionCtrl.Get)
				suggestions.POST("/:id/handle", middleware.RoleMiddleware(models.RoleKeeper, models.RoleAdmin), suggestionCtrl.Handle)
			}
		}
	}

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
