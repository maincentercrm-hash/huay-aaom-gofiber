package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupGenericRoutes(app *fiber.App, collection *mongo.Collection, searchFields, sortFields []string) {
	controller := controllers.NewGenericController(collection, searchFields, sortFields)

	group := app.Group("/api/" + collection.Name())
	group.Post("/", controller.Create)
	group.Get("/", controller.GetAll)
	group.Get("/search", controller.Search)
	group.Get("/:id", controller.GetById)
	group.Put("/:id", controller.Update)
	group.Delete("/:id", controller.Delete)
}
