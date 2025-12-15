package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupUserBetRoutes(app *fiber.App, db *mongo.Database) {
	collection := db.Collection("user_bets")
	configCollection := db.Collection("tbl_config")
	controller := controllers.NewUserBetController(collection, configCollection)

	userBetRoutes := app.Group("/api/user-bet")
	userBetRoutes.Get("/", controller.GetCurrentBet)
	userBetRoutes.Put("/", controller.UpdateCurrentBet)
}
