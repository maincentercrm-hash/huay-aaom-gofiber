package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-server/models"
	"go-server/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MissionController struct {
	missionCollection  *mongo.Collection
	configCollection   *mongo.Collection
	eventCollection    *mongo.Collection
	logCollection      *mongo.Collection // เพิ่ม logCollection
	telegramController *TelegramController
	lineController     *LineController
}

func NewMissionController(missionCollection, configCollection, eventCollection, logCollection *mongo.Collection, lineController *LineController) *MissionController {
	return &MissionController{
		missionCollection:  missionCollection,
		configCollection:   configCollection,
		eventCollection:    eventCollection,
		logCollection:      logCollection,
		telegramController: NewTelegramController(configCollection),
		lineController:     lineController,
	}
}

func (c *MissionController) GetProcessingMission(ctx *fiber.Ctx) error {
	userID := ctx.Query("user_id")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User ID is required"})
	}

	var mission models.Mission
	cursor, err := c.missionCollection.Find(
		ctx.Context(),
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(1),
	)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch mission"})
	}

	defer cursor.Close(ctx.Context())

	if cursor.Next(ctx.Context()) {
		if err := cursor.Decode(&mission); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode mission"})
		}
	} else {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No processing mission found"})
	}

	return ctx.JSON(mission)
}

func (c *MissionController) CreateMission(ctx *fiber.Ctx) error {
	// แก้ไขการรับ request ให้รับเฉพาะ field ที่ต้องการ
	var requestBody struct {
		UserID      string `json:"user_id"`
		PhoneNumber string `json:"phone_number"`
	}

	if err := ctx.BodyParser(&requestBody); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// ตรวจสอบว่ามี phone_number
	if requestBody.PhoneNumber == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Phone number is required"})
	}

	var config models.Config
	err := c.configCollection.FindOne(ctx.Context(), bson.M{}).Decode(&config)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch config"})
	}

	// สร้าง mission object จาก request body
	mission := &models.Mission{
		UserID:           requestBody.UserID,
		PhoneNumber:      requestBody.PhoneNumber,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Status:           "processing",
		CurrentTier:      1,
		ConsecutiveFails: 0,
	}

	if len(config.Tiers) > 0 {
		firstTierConfig := config.Tiers[0]
		mission.Tiers = []models.Tier{
			{
				Name:         firstTierConfig.Name,
				Reward:       firstTierConfig.Reward,
				Target:       firstTierConfig.Target,
				Status:       "processing",
				CurrentLevel: 1,
				MaxLevel:     firstTierConfig.MaxLevel,
				Levels: []models.Level{
					createNewLevel(1, firstTierConfig.Period, firstTierConfig.FollowUpHours),
				},
			},
		}
	} else {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No tier configuration found"})
	}

	result, err := c.missionCollection.InsertOne(ctx.Context(), mission)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create mission"})
	}

	mission.ID = result.InsertedID.(primitive.ObjectID)

	c.createNewEvents(ctx.Context(), mission, &mission.Tiers[0], &mission.Tiers[0].Levels[0], config.Tiers[0])

	return ctx.Status(fiber.StatusCreated).JSON(mission)
}

func (c *MissionController) UpdateMissionStatus(ctx *fiber.Ctx) error {
	missionID, err := primitive.ObjectIDFromHex(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid mission ID"})
	}

	var updateData struct {
		TierIndex int `json:"tierIndex"`
	}
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var mission models.Mission
	err = c.missionCollection.FindOne(ctx.Context(), bson.M{"_id": missionID}).Decode(&mission)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mission not found"})
	}

	var config models.Config
	err = c.configCollection.FindOne(ctx.Context(), bson.M{}).Decode(&config)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch config"})
	}

	currentTier := &mission.Tiers[updateData.TierIndex]
	currentLevel := &currentTier.Levels[currentTier.CurrentLevel-1]
	currentTierConfig := config.Tiers[updateData.TierIndex]

	// Get current bet for the mission
	currentBet, err := utils.GetCurrentBet(config, mission.UserID, currentLevel.StartDate, currentLevel.ExpireDate)
	if err != nil {
		log.Printf("Failed to get current bet: %v", err)
		currentBet = 0 // Fallback to 0 if API call fails
	}

	// Update current bet for the current level
	currentLevel.CurrentBet = currentBet

	if currentBet >= float64(currentTierConfig.Target) {
		if updateData.TierIndex == 2 {
			// Tier 3 handling
			log.Printf("Tier 3 level %d completed! Reward: %d", currentTier.CurrentLevel, currentTier.Reward)
			currentLevel.Status = "success"
			mission.ConsecutiveFails = 0 // Reset consecutive fails on success
			currentTier.CurrentLevel++
			newLevel := createNewLevel(currentTier.CurrentLevel, currentTierConfig.Period, currentTierConfig.FollowUpHours)
			currentTier.Levels = append(currentTier.Levels, newLevel)

			c.createNewEvents(ctx.Context(), &mission, currentTier, &newLevel, currentTierConfig)
		} else {
			// Tier 1 and 2 handling
			currentLevel.Status = "success"
			if currentTier.CurrentLevel < currentTier.MaxLevel {
				// Normal tier handling (not the last level)
				currentTier.CurrentLevel++
				newLevel := createNewLevel(currentTier.CurrentLevel, currentTierConfig.Period, currentTierConfig.FollowUpHours)
				currentTier.Levels = append(currentTier.Levels, newLevel)

				c.createNewEvents(ctx.Context(), &mission, currentTier, &newLevel, currentTierConfig)
			} else {
				// Last level of current tier
				currentTier.Status = "completed"
				log.Printf("Tier %d completed! Reward: %d", updateData.TierIndex+1, currentTier.Reward)

				if updateData.TierIndex < len(config.Tiers)-1 {
					// Move to next tier
					mission.CurrentTier++
					newTierConfig := config.Tiers[mission.CurrentTier-1]
					newTier := createNewTier(newTierConfig)
					mission.Tiers = append(mission.Tiers, newTier)
					log.Printf("Creating new Tier %d, starting at level 1", mission.CurrentTier)

					c.createNewEvents(ctx.Context(), &mission, &newTier, &newTier.Levels[0], newTierConfig)
				} else {
					mission.Status = "completed"
					log.Println("Mission completed")
				}
			}
		}
	} else {
		// Handle case when currentBet is less than target
		currentLevel.Status = "failed"
		mission.ConsecutiveFails++
		if mission.ConsecutiveFails >= currentTierConfig.MaxConsecutiveFails {
			mission.Status = "failed"
			log.Printf("Mission ID: %s - FAILED due to consecutive failures", mission.ID.Hex())
		}
	}

	mission.UpdatedAt = time.Now()

	_, err = c.missionCollection.UpdateOne(
		ctx.Context(),
		bson.M{"_id": missionID},
		bson.M{"$set": mission},
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update mission"})
	}

	return ctx.JSON(mission)
}

func (c *MissionController) ClaimReward(ctx *fiber.Ctx) error {
	missionID, err := primitive.ObjectIDFromHex(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid mission ID"})
	}

	var mission models.Mission
	err = c.missionCollection.FindOne(ctx.Context(), bson.M{"_id": missionID}).Decode(&mission)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mission not found"})
	}

	currentTier := &mission.Tiers[mission.CurrentTier-1]
	if currentTier.Status != "completed" && currentTier.Status != "awaiting_reward" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Current tier is not eligible for reward"})
	}

	// คำนวณ totalCurrentBet
	var totalCurrentBet float64
	for _, level := range currentTier.Levels {
		totalCurrentBet += level.CurrentBet
	}

	// Clear all reward-related events
	err = c.clearRewardRelatedEvents(ctx.Context(), missionID)
	if err != nil {
		log.Printf("Failed to clear reward related events: %v", err)
		// Continue processing even if clearing events fails
	}

	// Set mission status to pending
	_, err = c.missionCollection.UpdateOne(
		ctx.Context(),
		bson.M{"_id": missionID},
		bson.M{"$set": bson.M{
			"status": "pending",
			fmt.Sprintf("tiers.%d.status", mission.CurrentTier-1): "pending",
		}},
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update mission status"})
	}

	var config models.Config
	err = c.configCollection.FindOne(ctx.Context(), bson.M{}).Decode(&config)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch config"})
	}

	// Send Telegram message for claiming reward
	err = c.telegramController.SendRewardClaimedMessage(mission.ID.Hex(), mission.UserID, mission.CurrentTier, currentTier.CurrentLevel, currentTier.Reward)
	if err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
		// Continue with the process even if sending the message fails
	}

	// Send get reward notification
	err = c.lineController.SendGetRewardFlexMessage(
		mission.UserID,
		strconv.Itoa(mission.CurrentTier),
		strconv.Itoa(currentTier.CurrentLevel),
		mission.ID,
	)
	if err != nil {
		log.Printf("Failed to send get reward notification: %v", err)
	}

	// Prepare mission detail with totalCurrentBet
	var totalBet float64
	var startDate, endDate time.Time

	// หาวันที่เริ่มต้นและสิ้นสุดของ Tier
	for i, level := range currentTier.Levels {
		if i == 0 || level.StartDate.Before(startDate) {
			startDate = level.StartDate
		}
		if i == 0 || level.ExpireDate.After(endDate) {
			endDate = level.ExpireDate
		}
		totalBet += level.CurrentBet
	}

	// ปรับเวลาให้เป็น +7 สำหรับประเทศไทย
	thailandLoc, _ := time.LoadLocation("Asia/Bangkok")
	startDateTH := startDate.In(thailandLoc)
	endDateTH := endDate.In(thailandLoc)

	// สร้าง mission detail string
	missionDetailStr := fmt.Sprintf("Tier %d Complete ตั้งแต่วันที่ %s - %s รวมยอดเดิมพันทั้งสิ้น %s",
		mission.CurrentTier,
		startDateTH.Format("02/01/2006 15:04"),
		endDateTH.Format("02/01/2006 15:04"),
		formatNumber(totalBet))

	// ใช้ค่า reward โดยตรง
	rewardFloat := float64(currentTier.Reward)

	// Send reward claim to external API
	err = c.sendRewardClaimToExternalAPI(mission.UserID, rewardFloat, missionID.Hex(), missionDetailStr, &config)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send reward claim"})
	}

	return ctx.JSON(fiber.Map{"message": "Reward claim sent successfully", "status": "pending"})
}

func formatNumber(n float64) string {
	return strconv.FormatFloat(n, 'f', 2, 64)
}

func (c *MissionController) clearRewardRelatedEvents(ctx context.Context, missionID primitive.ObjectID) error {
	_, err := c.eventCollection.DeleteMany(ctx, bson.M{
		"mission_id": missionID,
		"type":       bson.M{"$in": []string{"reward_notification", "recurring_reward_notification", "reward_expiration"}},
		"status":     "pending",
	})
	if err != nil {
		log.Printf("Failed to clear reward related events: %v", err)
		return err
	}
	log.Printf("Cleared reward related events for Mission ID: %s", missionID.Hex())
	return nil
}

// แก้ไขฟังก์ชัน sendRewardClaimToExternalAPI
func (c *MissionController) sendRewardClaimToExternalAPI(userID string, reward float64, missionID string, missionDetail string, config *models.Config) error {
	logEntry := models.Log{
		UserID:        userID,
		MissionID:     missionID,
		MissionDetail: missionDetail,
		Reward:        reward,
		CreatedAt:     time.Now(),
		Status:        "pending",
	}

	result, err := c.logCollection.InsertOne(context.Background(), logEntry)
	if err != nil {
		log.Printf("Failed to create log entry: %v", err)
		return err
	}

	logID := result.InsertedID.(primitive.ObjectID)

	log.Printf("Sending reward claim: UserID: %s, Reward: %.2f, MissionID: %s, LogID: %s", userID, reward, missionID, logID.Hex())

	externalAPIPayload := map[string]interface{}{
		"log_id":         logID.Hex(),
		"user_id":        userID,
		"mission_detail": missionDetail,
		"reward":         reward,
		"callback_url":   fmt.Sprintf("%s/api/missions/reward-callback", os.Getenv("BASE_URL")),
		"line_at":        config.LineAt,
	}

	jsonData, err := json.Marshal(externalAPIPayload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", config.ApiEndpoint+"/players/v1/line/rewards/claim", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.ApiKey)

	log.Printf("Sending request to external API: %s", config.ApiEndpoint+"/players/v1/line/rewards/claim")
	log.Printf("Request payload: %s", string(jsonData))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request to external API: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return err
	}

	log.Printf("Response from external API - Status: %d, Body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("external API returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}

func (c *MissionController) createNewEvents(ctx context.Context, mission *models.Mission, currentTier *models.Tier, newLevel *models.Level, currentTierConfig models.TierDetail) {
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

	// Create reward_expiration event only when the tier is completed for Tier 1 and 2
	if mission.CurrentTier < 3 && currentTier.Status == "completed" {
		rewardExpirationEvent := models.ExpirationEvent{
			MissionID:  mission.ID,
			TierIndex:  mission.CurrentTier - 1,
			LevelIndex: currentTier.CurrentLevel - 1,
			ExpireTime: time.Now().Add(time.Duration(currentTierConfig.ExpireRewardHours) * time.Hour),
			Status:     "pending",
			Type:       "reward_expiration",
		}
		_, err = c.eventCollection.InsertOne(ctx, rewardExpirationEvent)
		if err != nil {
			log.Printf("Failed to create reward expiration event: %v", err)
		}
	}
}

func (c *MissionController) CheckExistingMission(ctx *fiber.Ctx) error {
	userID := ctx.Query("user_id")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User ID is required"})
	}

	// ค้นหา mission ที่มีสถานะ processing หรือ pending
	count, err := c.missionCollection.CountDocuments(ctx.Context(), bson.M{
		"user_id": userID,
		"status":  bson.M{"$in": []string{"processing", "pending"}},
	})

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check mission"})
	}

	return ctx.JSON(fiber.Map{
		"hasMission": count > 0,
	})
}
