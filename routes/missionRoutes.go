package routes

import (
	"go-server/controllers"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupMissionRoutes(app *fiber.App, db *mongo.Database) {
	missionCollection := db.Collection("tbl_mission")
	configCollection := db.Collection("tbl_config")
	eventCollection := db.Collection("tbl_events")
	logCollection := db.Collection("tbl_logs")
	messageCollection := db.Collection("tbl_logs_message")

	// สร้าง LineController
	lineController, err := controllers.NewLineController(configCollection, messageCollection)
	if err != nil {
		log.Fatal("Failed to create LINE controller:", err)
	}

	missionController := controllers.NewMissionController(missionCollection, configCollection, eventCollection, logCollection, lineController)
	rewardCallbackController := controllers.NewRewardCallbackController(missionCollection, logCollection, configCollection, eventCollection, lineController)

	missionRoutes := app.Group("/api/missions")
	missionRoutes.Post("/", missionController.CreateMission)
	missionRoutes.Put("/:id/status", missionController.UpdateMissionStatus)
	missionRoutes.Get("/processing", missionController.GetProcessingMission)
	missionRoutes.Post("/:id/claim-reward", missionController.ClaimReward)

	// เช็คว่าเคยกดรับหรือยัง
	missionRoutes.Get("/check", missionController.CheckExistingMission)

	// เพิ่ม route สำหรับ reward callback
	missionRoutes.Post("/reward-callback", rewardCallbackController.HandleRewardCallback)

}
