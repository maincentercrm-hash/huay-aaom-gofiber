package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupAdminRoutes(app *fiber.App, db *mongo.Database) {
	adminHandler := controllers.NewAdminHandler(db.Collection("users"))

	adminGroup := app.Group("/api/admin")
	adminGroup.Post("/login", adminHandler.Login)
	adminGroup.Post("/register", adminHandler.Register)
}
