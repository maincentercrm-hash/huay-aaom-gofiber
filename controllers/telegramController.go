package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-server/models"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TelegramController struct {
	configCollection *mongo.Collection
}

func NewTelegramController(configCollection *mongo.Collection) *TelegramController {
	return &TelegramController{
		configCollection: configCollection,
	}
}

// Message templates
var telegramMessages = struct {
	RewardClaimed func(missionID string, userId string, tier int, level int, reward int) string
}{
	RewardClaimed: func(missionID string, userId string, tier int, level int, reward int) string {
		return fmt.Sprintf(
			"<b>Request Reward Claimed!</b>\n\n"+
				"Mission ID: <code>%s</code>\n"+
				"User ID: <code>%s</code>\n"+
				"Tier: <b>%d</b>\n"+
				"Level: <b>%d</b>\n"+
				"Reward: <b>%d</b>",
			missionID, userId, tier, level, reward)
	},
}

func (tc *TelegramController) SendRewardClaimedMessage(missionID string, userId string, tier int, level int, reward int) error {
	var config models.Config
	err := tc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	botToken := config.TelegramBotToken
	chatID := config.TelegramChatID

	message := telegramMessages.RewardClaimed(missionID, userId, tier, level, reward)

	return tc.sendHTMLMessage(botToken, chatID, message)
}

func (tc *TelegramController) sendHTMLMessage(botToken, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	body, err := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Description string `json:"description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return fmt.Errorf("unexpected status code: %d, error: %s", resp.StatusCode, errorResponse.Description)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
