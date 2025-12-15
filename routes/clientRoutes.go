package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupClientRoutes(app *fiber.App, db *mongo.Database) {
	clientCollection := db.Collection("tbl_client")
	configCollection := db.Collection("tbl_config")
	clientController := controllers.NewClientController(clientCollection, configCollection)

	clientGroup := app.Group("/api/clients")
	clientGroup.Get("/", clientController.GetAllClients)
	clientGroup.Post("/", clientController.UpsertClient)
	clientGroup.Get("/:userId", clientController.GetClientByUserId)
	clientGroup.Delete("/:userId", clientController.DeleteClient)
	clientGroup.Get("/:userId/check-phone", clientController.CheckPhoneNumber)
	clientGroup.Put("/:userId/update-phone", clientController.UpdatePhoneNumber)
}
