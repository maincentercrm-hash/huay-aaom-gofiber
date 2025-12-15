package controllers

import (
	"context"
	"log"
	"time"

	"go-server/models"
	"go-server/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserBetController struct {
	collection       *mongo.Collection
	configCollection *mongo.Collection
}

func NewUserBetController(collection, configCollection *mongo.Collection) *UserBetController {
	return &UserBetController{
		collection:       collection,
		configCollection: configCollection,
	}
}

func (c *UserBetController) GetCurrentBet(ctx *fiber.Ctx) error {
	log.Println("GetCurrentBet: Starting")
	userID := ctx.Query("userId")
	startDateStr := ctx.Query("startDate")
	endDateStr := ctx.Query("endDate")

	log.Printf("GetCurrentBet: Received params - userID: %s, startDate: %s, endDate: %s", userID, startDateStr, endDateStr)

	if userID == "" || startDateStr == "" || endDateStr == "" {
		log.Println("GetCurrentBet: Missing required parameters")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User ID, start date, and end date are required"})
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		log.Printf("GetCurrentBet: Invalid start date format - %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date format"})
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		log.Printf("GetCurrentBet: Invalid end date format - %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid end date format"})
	}

	// Return a fixed value of 500 for currentBet
	/*
		currentBet := 500.0
		log.Printf("GetCurrentBet: Returning fixed currentBet value of %.2f", currentBet)
		return ctx.JSON(fiber.Map{"bet": currentBet})
	*/

	log.Println("GetCurrentBet: Fetching config")
	var config models.Config
	err = c.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		log.Printf("GetCurrentBet: Failed to fetch config - %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	log.Println("GetCurrentBet: Calling utils.GetCurrentBet")
	currentBet, err := utils.GetCurrentBet(config, userID, startDate, endDate)
	if err != nil {
		log.Printf("GetCurrentBet: Failed to get current bet - %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get current bet"})
	}

	log.Printf("GetCurrentBet: Successfully retrieved current bet - %.2f", currentBet)
	return ctx.JSON(fiber.Map{"bet": currentBet})

}

func (c *UserBetController) UpdateCurrentBet(ctx *fiber.Ctx) error {
	var input struct {
		UserID     string  `json:"userId"`
		CurrentBet float64 `json:"currentBet"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	update := bson.M{
		"$set": bson.M{
			"current_bet": input.CurrentBet,
			"updated_at":  time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := c.collection.UpdateOne(
		context.Background(),
		bson.M{"user_id": input.UserID},
		update,
		opts,
	)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update current bet"})
	}

	return ctx.JSON(fiber.Map{"message": "Current bet updated successfully"})
}
