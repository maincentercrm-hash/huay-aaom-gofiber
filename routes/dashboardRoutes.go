// routes/dashboard_routes.go
package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupDashboardRoutes(app *fiber.App, db *mongo.Database) {
	dashboardController := controllers.NewDashboardController(
		db.Collection("tbl_mission"),
		db.Collection("tbl_client"),
		db.Collection("tbl_user_bet"),
		db.Collection("tbl_logs"),
	)

	dashboardGroup := app.Group("/api/dashboard")
	dashboardGroup.Get("/", dashboardController.GetDashboardData)
	dashboardGroup.Get("/stats", dashboardController.GetStatsData)
	dashboardGroup.Get("/tier-performance", dashboardController.GetTierPerformanceData)
	dashboardGroup.Get("/urgent-alerts", dashboardController.GetUrgentAlertsData)
	dashboardGroup.Get("/recent-activities", dashboardController.GetRecentActivitiesData)
	dashboardGroup.Get("/pending-rewards", dashboardController.GetPendingRewards)
}
