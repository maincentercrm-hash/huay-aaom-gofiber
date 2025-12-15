package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupConfigRoutes(app *fiber.App, db *mongo.Database) {
	configCollection := db.Collection("tbl_config")
	configController := controllers.NewConfigController(configCollection)

	configRoutes := app.Group("/api/config")
	configRoutes.Get("/", configController.GetConfig)
	configRoutes.Post("/", configController.SaveConfig)
	configRoutes.Put("/tiers", configController.UpdateTierSettings)
	configRoutes.Put("/flex-messages", configController.UpdateFlexMessageSettings)
	configRoutes.Put("/site-template", configController.UpdateSiteTemplateConfig)
	configRoutes.Post("/upload-image", configController.UploadImage)
}
