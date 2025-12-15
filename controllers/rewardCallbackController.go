package controllers

import (
	"context"
	"fmt"
	"go-server/models"
	"go-server/utils"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RewardCallbackController struct {
	missionCollection *mongo.Collection
	logCollection     *mongo.Collection
	configCollection  *mongo.Collection
	eventCollection   *mongo.Collection
	lineController    *LineController
}

func NewRewardCallbackController(missionCollection, logCollection, configCollection, eventCollection *mongo.Collection, lineController *LineController) *RewardCallbackController {
	return &RewardCallbackController{
		missionCollection: missionCollection,
		logCollection:     logCollection,
		configCollection:  configCollection,
		eventCollection:   eventCollection,
		lineController:    lineController,
	}
}

func (c *RewardCallbackController) HandleRewardCallback(ctx *fiber.Ctx) error {
	var callback struct {
		LogID  string `json:"log_id"`
		Status string `json:"status"` // "approve" or "reject"
	}
	if err := ctx.BodyParser(&callback); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	logID, err := primitive.ObjectIDFromHex(callback.LogID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid log ID"})
	}

	// Fetch the log entry to get the mission_id
	var logEntry models.Log
	err = c.logCollection.FindOne(context.Background(), bson.M{"_id": logID}).Decode(&logEntry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Log entry not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch log entry"})
	}

	callbackTime := time.Now()
	// Update the existing log
	update := bson.M{
		"$set": bson.M{
			"callback_time": callbackTime,
			"status":        callback.Status,
		},
	}

	result, err := c.logCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": logID, "status": "pending"},
		update,
	)

	if err != nil {
		log.Printf("Failed to update log: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process callback"})
	}

	if result.MatchedCount == 0 {
		log.Printf("No pending log found for LogID: %s", callback.LogID)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No matching pending reward claim found"})
	}

	missionID, err := primitive.ObjectIDFromHex(logEntry.MissionID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid mission ID in log entry"})
	}

	var mission models.Mission
	err = c.missionCollection.FindOne(ctx.Context(), bson.M{"_id": missionID}).Decode(&mission)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mission not found"})
	}

	if callback.Status == "approve" {
		err = c.processSuccessfulReward(ctx.Context(), &mission)
	} else if callback.Status == "reject" {
		err = c.processFailedReward(ctx.Context(), &mission)
	} else {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status. Must be 'approve' or 'reject'"})
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("Reward callback processed successfully for Mission ID: %s, New Tier: %d, New Level: %d",
		mission.ID.Hex(), mission.CurrentTier, mission.Tiers[mission.CurrentTier-1].CurrentLevel)

	// Prepare the new response
	var message string
	if callback.Status == "approve" {
		message = "Reward approved successfully"
	} else {
		message = "Reward rejected"
	}

	response := fiber.Map{
		"log_id":        logID.Hex(),
		"callback_time": callbackTime,
		"status":        callback.Status,
		"message":       message,
	}

	// Return both the original success message and the new response
	return ctx.JSON(response)
}

func (c *RewardCallbackController) processSuccessfulReward(ctx context.Context, mission *models.Mission) error {
	currentTier := &mission.Tiers[mission.CurrentTier-1]
	currentTier.Status = "completed"

	var config models.Config
	err := c.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	// Get current bet for the mission
	currentLevel := currentTier.Levels[currentTier.CurrentLevel-1]
	currentBet, err := utils.GetCurrentBet(config, mission.UserID, currentLevel.StartDate, currentLevel.ExpireDate)
	if err != nil {
		log.Printf("Failed to get current bet: %v", err)
		currentBet = 0 // Fallback to 0 if API call fails
	}

	if mission.CurrentTier < 3 {
		// Move to next tier for Tier 1 and 2
		mission.CurrentTier++
		newTierConfig := config.Tiers[mission.CurrentTier-1]
		newTier := createNewTier(newTierConfig)
		newTier.Levels[0].CurrentBet = currentBet // Set current bet for the first level of new tier
		mission.Tiers = append(mission.Tiers, newTier)

		// Create new events for the first level of the new tier
		c.createNewEvents(ctx, mission, &newTier, &newTier.Levels[0], newTierConfig)
	} else {
		// For Tier 3, create a new level
		currentTierConfig := config.Tiers[mission.CurrentTier-1]
		currentTier.CurrentLevel++
		newLevel := createNewLevel(currentTier.CurrentLevel, currentTierConfig.Period, currentTierConfig.FollowUpHours)
		newLevel.CurrentBet = currentBet // Set current bet for the new level
		currentTier.Levels = append(currentTier.Levels, newLevel)

		// Create new events for the new level in Tier 3
		c.createNewEvents(ctx, mission, currentTier, &newLevel, currentTierConfig)
	}

	mission.Status = "processing"
	currentTier.Status = "processing"
	mission.ConsecutiveFails = 0
	mission.UpdatedAt = time.Now()

	_, err = c.missionCollection.UpdateOne(
		ctx,
		bson.M{"_id": mission.ID},
		bson.M{"$set": mission},
	)
	if err != nil {
		return fmt.Errorf("failed to update mission: %v", err)
	}

	log.Printf("Successfully processed reward for Mission ID: %s, New Tier: %d, New Level: %d, Current Bet: %.2f",
		mission.ID.Hex(), mission.CurrentTier, mission.Tiers[mission.CurrentTier-1].CurrentLevel, currentBet)

	return nil
}

func (c *RewardCallbackController) processFailedReward(ctx context.Context, mission *models.Mission) error {
	currentTier := &mission.Tiers[mission.CurrentTier-1]
	currentTier.Status = "awaiting_reward" // กลับไปสู่สถานะรอรับรางวัล
	mission.Status = "processing"
	mission.UpdatedAt = time.Now()

	_, err := c.missionCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": mission.ID},
		bson.M{"$set": mission},
	)
	return err
}

func (c *RewardCallbackController) createNewEvents(ctx context.Context, mission *models.Mission, currentTier *models.Tier, newLevel *models.Level, currentTierConfig models.TierDetail) {
	processingDelay := time.Duration(currentTierConfig.ProcessingDelay) * time.Minute
	log.Printf("Creating new events with Processing Delay: %v", processingDelay)

	levelExpirationEvent := models.ExpirationEvent{
		MissionID:  mission.ID,
		TierIndex:  mission.CurrentTier - 1,
		LevelIndex: currentTier.CurrentLevel - 1,
		ExpireTime: newLevel.ExpireDate.Add(processingDelay),
		Status:     "pending",
		Type:       "level_expiration",
	}
	_, err := c.eventCollection.InsertOne(ctx, levelExpirationEvent)
	if err != nil {
		log.Printf("Failed to create new level expiration event: %v", err)
	}
	log.Printf("Level expiration event created with ExpireTime: %v", levelExpirationEvent.ExpireTime)

	followUpEvent := models.ExpirationEvent{
		MissionID:  mission.ID,
		TierIndex:  mission.CurrentTier - 1,
		LevelIndex: currentTier.CurrentLevel - 1,
		ExpireTime: newLevel.FollowUpDate,
		Status:     "pending",
		Type:       "follow_up",
	}
	_, err = c.eventCollection.InsertOne(ctx, followUpEvent)
	if err != nil {
		log.Printf("Failed to create follow-up event: %v", err)
	}

	log.Printf("Created new events for Mission ID: %s, Tier: %d, Level: %d", mission.ID.Hex(), mission.CurrentTier, currentTier.CurrentLevel)
}

// Helper functions (createNewTier and createNewLevel) should be defined here or imported from another package
