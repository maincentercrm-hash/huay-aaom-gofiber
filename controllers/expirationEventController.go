package controllers

import (
	"context"
	"fmt"
	"go-server/models"
	"log"
	"strconv"
	"time"

	"go-server/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExpirationEventController struct {
	eventCollection   *mongo.Collection
	missionCollection *mongo.Collection
	configCollection  *mongo.Collection
	lineController    *LineController
}

func NewExpirationEventController(eventCollection, missionCollection, configCollection *mongo.Collection, lineController *LineController) *ExpirationEventController {
	return &ExpirationEventController{
		eventCollection:   eventCollection,
		missionCollection: missionCollection,
		configCollection:  configCollection,
		lineController:    lineController,
	}
}

func (c *ExpirationEventController) ProcessEvents() {
	log.Println("Starting ProcessEvents")
	for {
		ctx := context.Background()
		now := time.Now()
		filter := bson.M{
			"expire_time": bson.M{"$lte": now},
			"status":      "pending",
		}
		update := bson.M{"$set": bson.M{"status": "processed"}}
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

		var event models.ExpirationEvent
		err := c.eventCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&event)
		if err == mongo.ErrNoDocuments {
			//	log.Println("No pending events found, sleeping for 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
		if err != nil {
			log.Printf("Error processing expiration event: %v", err)
			continue
		}

		log.Printf("Processing event: Type: %s, MissionID: %s", event.Type, event.MissionID.Hex())
		err = c.handleExpiredMission(ctx, event)
		if err != nil {
			log.Printf("Error handling expired mission: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (c *ExpirationEventController) handleExpiredMission(ctx context.Context, event models.ExpirationEvent) error {
	var mission models.Mission
	err := c.missionCollection.FindOne(ctx, bson.M{"_id": event.MissionID}).Decode(&mission)
	if err != nil {
		return err
	}

	var config models.Config
	err = c.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		return err
	}

	currentTier := &mission.Tiers[event.TierIndex]
	currentLevel := &currentTier.Levels[event.LevelIndex]
	currentTierConfig := config.Tiers[event.TierIndex]

	switch event.Type {
	case "level_expiration":
		return c.handleLevelExpiration(ctx, &mission, currentTier, currentLevel, currentTierConfig)
	case "follow_up":
		return c.handleFollowUp(ctx, &mission, currentLevel, currentTierConfig)
	case "reward_expiration":
		return c.handleRewardExpiration(ctx, &mission, currentTier, currentTierConfig)
	case "reward_notification", "recurring_reward_notification":
		return c.handleRewardNotification(ctx, &mission, currentTier, currentTierConfig)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (c *ExpirationEventController) handleLevelExpiration(ctx context.Context, mission *models.Mission, currentTier *models.Tier, currentLevel *models.Level, currentTierConfig models.TierDetail) error {
	var config models.Config
	err := c.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	// Calculate the actual expiration time without processing delay
	actualExpireTime := currentLevel.ExpireDate

	currentBet, err := utils.GetCurrentBet(config, mission.UserID, currentLevel.StartDate, actualExpireTime)
	if err != nil {
		log.Printf("Failed to get current bet: %v", err)
		currentBet = 0 // Fallback to 0 if API call fails
	}

	currentLevel.CurrentBet = currentBet

	if currentBet >= float64(currentTierConfig.Target) {
		currentLevel.Status = "success"
		log.Printf("Mission ID: %s, Tier: %d, Level: %d - SUCCESS", mission.ID.Hex(), mission.CurrentTier, currentTier.CurrentLevel)

		mission.ConsecutiveFails = 0 // Reset consecutive fails on success

		if mission.CurrentTier < 3 {
			// Tier 1 or 2 handling
			if currentTier.CurrentLevel == currentTier.MaxLevel {
				currentTier.Status = "awaiting_reward"
				log.Printf("Mission ID: %s, Tier: %d - COMPLETE, AWAITING REWARD", mission.ID.Hex(), mission.CurrentTier)
				c.createRewardExpirationEvent(ctx, mission, currentTier, currentTierConfig)

				err := c.lineController.SendMissionCompleteFlexMessage(
					mission.UserID,
					fmt.Sprintf("%d", currentTierConfig.ExpireRewardHours),
					strconv.Itoa(mission.CurrentTier),
					strconv.Itoa(currentTier.CurrentLevel),
					mission.ID,
				)
				if err != nil {
					log.Printf("Failed to send mission complete notification: %v", err)
				}
			} else {
				// Prepare next level for Tier 1 and 2
				currentTier.CurrentLevel++
				newLevel := createNewLevel(currentTier.CurrentLevel, currentTierConfig.Period, currentTierConfig.FollowUpHours)
				currentTier.Levels = append(currentTier.Levels, newLevel)
				c.createNewEvents(ctx, mission, currentTier, &newLevel, currentTierConfig)

				err := c.lineController.SendMissionSuccessFlexMessage(
					mission.UserID,
					strconv.Itoa(mission.CurrentTier),
					strconv.Itoa(currentTier.CurrentLevel-1), // Send the completed level
					mission.ID,
				)
				if err != nil {
					log.Printf("Failed to send mission success notification: %v", err)
				}
			}
		} else {
			// Tier 3 handling
			currentTier.Status = "awaiting_reward"
			log.Printf("Mission ID: %s, Tier: 3, Level: %d - AWAITING REWARD", mission.ID.Hex(), currentTier.CurrentLevel)
			c.createRewardExpirationEvent(ctx, mission, currentTier, currentTierConfig)

			err := c.lineController.SendMissionCompleteFlexMessage(
				mission.UserID,
				fmt.Sprintf("%d", currentTierConfig.ExpireRewardHours),
				strconv.Itoa(mission.CurrentTier),
				strconv.Itoa(currentTier.CurrentLevel),
				mission.ID,
			)
			if err != nil {
				log.Printf("Failed to send mission complete notification: %v", err)
			}
		}
	} else {
		currentLevel.Status = "failed"
		currentTier.Status = "failed" // Update tier status to failed
		log.Printf("Mission ID: %s, Tier: %d, Level: %d - FAILED", mission.ID.Hex(), mission.CurrentTier, currentTier.CurrentLevel)

		if mission.CurrentTier < 3 {
			mission.Status = "failed"
			log.Printf("Mission ID: %s - FAILED in Tier %d", mission.ID.Hex(), mission.CurrentTier)
		} else {
			mission.ConsecutiveFails++
			if mission.ConsecutiveFails >= currentTierConfig.MaxConsecutiveFails {
				mission.Status = "failed"
				log.Printf("Mission ID: %s - FAILED due to consecutive failures in Tier 3", mission.ID.Hex())
			} else {
				// Continue to next level in Tier 3
				currentTier.CurrentLevel++
				newLevel := createNewLevel(currentTier.CurrentLevel, currentTierConfig.Period, currentTierConfig.FollowUpHours)
				currentTier.Levels = append(currentTier.Levels, newLevel)
				c.createNewEvents(ctx, mission, currentTier, &newLevel, currentTierConfig)
				currentTier.Status = "processing"
			}
		}

		// Send mission failed notification
		err := c.lineController.SendMissionFailedFlexMessage(
			mission.UserID,
			fmt.Sprintf("%d", currentTierConfig.Target),
			strconv.Itoa(mission.CurrentTier),
			strconv.Itoa(currentTier.CurrentLevel),
			mission.ID,
		)
		if err != nil {
			log.Printf("Failed to send mission failed notification: %v", err)
		}
	}

	mission.UpdatedAt = time.Now()
	_, err = c.missionCollection.UpdateOne(
		ctx,
		bson.M{"_id": mission.ID},
		bson.M{"$set": mission},
	)
	if err != nil {
		return fmt.Errorf("failed to update mission: %v", err)
	}

	log.Printf("Mission ID: %s updated. Status: %s, Tier Status: %s, Level Status: %s",
		mission.ID.Hex(), mission.Status, currentTier.Status, currentLevel.Status)

	return nil
}

func (c *ExpirationEventController) handleFollowUp(ctx context.Context, mission *models.Mission, currentLevel *models.Level, currentTierConfig models.TierDetail) error {
	var config models.Config
	err := c.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	currentBet, err := utils.GetCurrentBet(config, mission.UserID, currentLevel.StartDate, currentLevel.ExpireDate)
	if err != nil {
		log.Printf("Failed to get current bet: %v", err)
		currentBet = 0 // Fallback to 0 if API call fails
	}

	log.Printf("Follow-up for Mission ID %s, Tier %d, Level %s. Current bet: %.2f, Target: %d",
		mission.ID.Hex(), mission.CurrentTier, currentLevel.Name, currentBet, currentTierConfig.Target)

	// Convert mission.CurrentTier (int) to string
	tierString := strconv.Itoa(mission.CurrentTier)

	// Convert currentLevel.Name (already a string) to the format expected (e.g., "level 1")
	levelString := currentLevel.Name

	// Send follow-up notification
	err = c.lineController.SendFollowUpFlexMessage(
		mission.UserID,
		fmt.Sprintf("%d", currentTierConfig.Target),
		fmt.Sprintf("%.2f", currentBet),
		tierString,
		levelString,
		mission.ID, // This is already a primitive.ObjectID, no conversion needed
	)
	if err != nil {
		log.Printf("Failed to send follow-up notification: %v", err)
	}

	return nil
}

func (c *ExpirationEventController) handleRewardExpiration(ctx context.Context, mission *models.Mission, currentTier *models.Tier, currentTierConfig models.TierDetail) error {
	var latestMission models.Mission
	err := c.missionCollection.FindOne(ctx, bson.M{"_id": mission.ID}).Decode(&latestMission)
	if err != nil {
		return err
	}

	if latestMission.CurrentTier == mission.CurrentTier &&
		(latestMission.Status == "processing" || latestMission.Tiers[mission.CurrentTier-1].Status == "awaiting_reward") {
		if (currentTier.Status == "completed" || currentTier.Status == "awaiting_reward") &&
			time.Now().After(currentTier.ExpireReward) {
			// Update tier status to expire_reward
			latestMission.Tiers[mission.CurrentTier-1].Status = "expire_reward"
			log.Printf("Mission ID: %s, Tier: %d - REWARD EXPIRED", mission.ID.Hex(), mission.CurrentTier)

			// Check if this is the last tier
			if mission.CurrentTier == len(latestMission.Tiers) {
				latestMission.Status = "failed"
			} else {
				// Move to the next tier if available
				latestMission.CurrentTier++
				latestMission.Status = "processing"
			}

			_, err := c.missionCollection.UpdateOne(
				ctx,
				bson.M{"_id": mission.ID},
				bson.M{"$set": bson.M{
					"status":       latestMission.Status,
					"current_tier": latestMission.CurrentTier,
					fmt.Sprintf("tiers.%d.status", mission.CurrentTier-1): "expire_reward",
				}},
			)
			if err != nil {
				log.Printf("Failed to update mission status after reward expiration: %v", err)
				return err
			}

			return nil
		}
	} else {
		log.Printf("Mission ID: %s - Reward expiration event skipped (mission state changed)", mission.ID.Hex())
	}
	return nil
}

func (c *ExpirationEventController) handleRewardNotification(ctx context.Context, mission *models.Mission, currentTier *models.Tier, currentTierConfig models.TierDetail) error {
	log.Printf("Sending reward notification for Mission ID: %s, Tier: %d", mission.ID.Hex(), mission.CurrentTier)

	remainingDays := int(time.Until(currentTier.ExpireReward).Hours() / 24)

	// Convert mission.CurrentTier (int) to string
	tierString := strconv.Itoa(mission.CurrentTier)

	// Convert currentTier.CurrentLevel (int) to string
	levelString := strconv.Itoa(currentTier.CurrentLevel)

	err := c.lineController.SendRewardNotificationFlexMessage(
		mission.UserID,
		fmt.Sprintf("%d", remainingDays),
		tierString,
		levelString,
		mission.ID, // This is already a primitive.ObjectID, no conversion needed
	)
	if err != nil {
		log.Printf("Failed to send reward notification: %v", err)
	}

	return nil
}

func (c *ExpirationEventController) createNewEvents(ctx context.Context, mission *models.Mission, currentTier *models.Tier, newLevel *models.Level, currentTierConfig models.TierDetail) {
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
}

func (c *ExpirationEventController) createRewardExpirationEvent(ctx context.Context, mission *models.Mission, currentTier *models.Tier, currentTierConfig models.TierDetail) {
	expireRewardTime := time.Now().Add(time.Duration(currentTierConfig.ExpireRewardHours) * time.Hour)
	currentTier.ExpireReward = expireRewardTime

	rewardExpirationEvent := models.ExpirationEvent{
		MissionID:  mission.ID,
		TierIndex:  mission.CurrentTier - 1,
		LevelIndex: currentTier.CurrentLevel - 1,
		ExpireTime: expireRewardTime,
		Status:     "pending",
		Type:       "reward_expiration",
	}
	_, err := c.eventCollection.InsertOne(ctx, rewardExpirationEvent)
	if err != nil {
		log.Printf("Failed to create reward expiration event: %v", err)
	}

	// Create notification events based on tier
	if mission.CurrentTier < 3 {
		// For Tier 1 and 2: Create a single notification event before expiration
		if currentTierConfig.NotifyBeforeExpire > 0 {
			notifyTime := expireRewardTime.Add(-time.Duration(currentTierConfig.NotifyBeforeExpire) * time.Hour)
			notificationEvent := models.ExpirationEvent{
				MissionID:  mission.ID,
				TierIndex:  mission.CurrentTier - 1,
				LevelIndex: currentTier.CurrentLevel - 1,
				ExpireTime: notifyTime,
				Status:     "pending",
				Type:       "reward_notification",
			}
			_, err := c.eventCollection.InsertOne(ctx, notificationEvent)
			if err != nil {
				log.Printf("Failed to create notification event for Tier %d: %v", mission.CurrentTier, err)
			}
		}
	} else {
		// For Tier 3: Create recurring notification events
		if currentTierConfig.NotifyInterval > 0 {
			now := time.Now()
			notifyInterval := time.Duration(currentTierConfig.NotifyInterval) * time.Hour
			for notifyTime := now.Add(notifyInterval); notifyTime.Before(expireRewardTime); notifyTime = notifyTime.Add(notifyInterval) {
				notificationEvent := models.ExpirationEvent{
					MissionID:  mission.ID,
					TierIndex:  mission.CurrentTier - 1,
					LevelIndex: currentTier.CurrentLevel - 1,
					ExpireTime: notifyTime,
					Status:     "pending",
					Type:       "recurring_reward_notification",
				}
				_, err := c.eventCollection.InsertOne(ctx, notificationEvent)
				if err != nil {
					log.Printf("Failed to create recurring notification event for Tier 3: %v", err)
				}
			}
		}
	}

	// Update the mission with the new ExpireReward time and status
	update := bson.M{
		fmt.Sprintf("tiers.%d.expire_reward", mission.CurrentTier-1): expireRewardTime,
		fmt.Sprintf("tiers.%d.status", mission.CurrentTier-1):        "completed",
	}

	if mission.CurrentTier == 3 {
		// For Tier 3, also update the current level's status
		update[fmt.Sprintf("tiers.%d.levels.%d.status", mission.CurrentTier-1, currentTier.CurrentLevel-1)] = "completed"
	}

	_, err = c.missionCollection.UpdateOne(
		ctx,
		bson.M{"_id": mission.ID},
		bson.M{"$set": update},
	)
	if err != nil {
		log.Printf("Failed to update mission with new ExpireReward time and status: %v", err)
	}
}

func (c *ExpirationEventController) getNextTierConfig(ctx context.Context, nextTierIndex int) models.TierDetail {
	var config models.Config
	err := c.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		log.Printf("Failed to fetch config: %v", err)
		return models.TierDetail{} // Return empty config in case of error
	}
	return config.Tiers[nextTierIndex-1]
}

func createNewLevel(levelNumber, period, followUpDays int) models.Level {
	now := time.Now()
	return models.Level{
		Name:         fmt.Sprintf("level %d", levelNumber),
		StartDate:    now,
		ExpireDate:   now.Add(time.Duration(period) * time.Hour),
		FollowUpDate: now.Add(time.Duration(followUpDays) * time.Hour),
		Status:       "processing",
		CurrentBet:   0,
	}
}

func createNewTier(tierConfig models.TierDetail) models.Tier {
	return models.Tier{
		Name:         tierConfig.Name,
		Reward:       tierConfig.Reward,
		Target:       tierConfig.Target,
		Status:       "processing",
		CurrentLevel: 1,
		MaxLevel:     tierConfig.MaxLevel,
		Levels: []models.Level{
			createNewLevel(1, tierConfig.Period, tierConfig.FollowUpHours),
		},
	}
}
