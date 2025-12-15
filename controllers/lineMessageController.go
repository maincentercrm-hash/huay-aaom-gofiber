package controllers

import (
	"context"
	"fmt"
	"go-server/models"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LineController struct {
	configCollection  *mongo.Collection
	messageCollection *mongo.Collection
	bot               *linebot.Client
}

func NewLineController(configCollection, messageCollection *mongo.Collection) (*LineController, error) {
	var config models.Config
	err := configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %v", err)
	}

	bot, err := linebot.New(config.ChannelSecret, config.ChannelAccessToken)
	if err != nil {
		return nil, fmt.Errorf("error creating LINE bot client: %v", err)
	}

	return &LineController{
		configCollection:  configCollection,
		messageCollection: messageCollection,
		bot:               bot,
	}, nil
}

func (lc *LineController) logMessage(userID string, tier string, level string, missionID primitive.ObjectID, flexConfig models.BaseFlexMessageContent, placeholders map[string]string) error {
	messageLog := models.MessageLog{
		UserID:    userID,
		Status:    "unread",
		Tier:      tier,
		Level:     level,
		MissionID: missionID,
		SentAt:    time.Now(),
		FlexContent: models.FlexContent{
			Title:          flexConfig.Title,
			Description:    replaceePlaceholders(flexConfig.Description, placeholders),
			SubDescription: replaceePlaceholders(flexConfig.SubDescription, placeholders),
		},
	}

	_, err := lc.messageCollection.InsertOne(context.Background(), messageLog)
	return err
}

func (lc *LineController) sendFlexMessageAndLog(userID, tier, level string, missionID primitive.ObjectID, flexConfig models.BaseFlexMessageContent, placeholders map[string]string) error {
	flexMessage := createFlexMessage(flexConfig, placeholders)
	_, err := lc.bot.PushMessage(userID, flexMessage).Do()

	logErr := lc.logMessage(userID, tier, level, missionID, flexConfig, placeholders)

	if err != nil && logErr == nil {
		return fmt.Errorf("message not sent (possibly due to quota), but logged successfully: %v", err)
	}

	if err != nil && logErr != nil {
		return fmt.Errorf("failed to send message: %v; failed to log message: %v", err, logErr)
	}

	if logErr != nil {
		return fmt.Errorf("message sent, but failed to log: %v", logErr)
	}

	return nil
}

func createFlexMessage(flexConfig models.BaseFlexMessageContent, placeholders map[string]string) *linebot.FlexMessage {
	// Replace placeholders in description and subDescription
	description := replaceePlaceholders(flexConfig.Description, placeholders)
	subDescription := replaceePlaceholders(flexConfig.SubDescription, placeholders)

	container := &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Hero: &linebot.ImageComponent{
			Type:        linebot.FlexComponentTypeImage,
			URL:         flexConfig.ImageUrl,
			Size:        linebot.FlexImageSizeTypeFull,
			AspectRatio: linebot.FlexImageAspectRatioType20to13,
			AspectMode:  linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   flexConfig.Title,
					Weight: linebot.FlexTextWeightTypeBold,
					Size:   linebot.FlexTextSizeTypeXl,
					Align:  linebot.FlexComponentAlignTypeCenter,
					Color:  "#27AE60", // Green color for title
				},
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  description,
					Wrap:  true,
					Align: linebot.FlexComponentAlignTypeCenter,
					Size:  linebot.FlexTextSizeTypeMd,
				},
			},
		},
	}

	if subDescription != "" {
		container.Body.Contents = append(container.Body.Contents, &linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   subDescription,
			Wrap:   true,
			Align:  linebot.FlexComponentAlignTypeCenter,
			Size:   linebot.FlexTextSizeTypeSm,
			Color:  "#888888",
			Margin: linebot.FlexComponentMarginTypeMd,
		})
	}

	if flexConfig.ButtonTitle != "" && flexConfig.ButtonUrl != "" {
		container.Footer = &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#27AE60", // Green color for button
					Action: linebot.NewURIAction(flexConfig.ButtonTitle, flexConfig.ButtonUrl),
					Height: linebot.FlexButtonHeightTypeMd,
				},
			},
		}
	}

	return linebot.NewFlexMessage(flexConfig.Title, container)
}

func replaceePlaceholders(text string, placeholders map[string]string) string {
	for key, value := range placeholders {
		// Try to convert the value to a float64
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			// If successful, format it as an integer with comma as thousand separator
			value = formatNumberWithCommas(int(floatValue))
		}
		text = strings.ReplaceAll(text, "{"+key+"}", value)
	}
	return text
}

func formatNumberWithCommas(n int) string {
	in := strconv.Itoa(n)
	numOfDigits := len(in)
	if numOfDigits < 4 {
		return in
	}
	out := make([]byte, 0, numOfDigits+(numOfDigits-1)/3)
	for i, c := range in {
		if i > 0 && (numOfDigits-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func (lc *LineController) SendFollowUpFlexMessage(userID, target, currentBet, tier, level string, missionID primitive.ObjectID) error {
	var config models.Config
	err := lc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	placeholders := map[string]string{
		"target":     target,
		"currentBet": currentBet,
	}

	return lc.sendFlexMessageAndLog(userID, tier, level, missionID, config.FlexMessages.Followup, placeholders)
}

func (lc *LineController) SendMissionSuccessFlexMessage(userID, tier, level string, missionID primitive.ObjectID) error {
	var config models.Config
	err := lc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	return lc.sendFlexMessageAndLog(userID, tier, level, missionID, config.FlexMessages.MissionSuccess, nil)
}

func (lc *LineController) SendMissionFailedFlexMessage(userID, target, tier, level string, missionID primitive.ObjectID) error {
	var config models.Config
	err := lc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	placeholders := map[string]string{
		"target": target,
	}

	return lc.sendFlexMessageAndLog(userID, tier, level, missionID, config.FlexMessages.MissionFailed, placeholders)
}

func (lc *LineController) SendMissionCompleteFlexMessage(userID, expireRewardDays, tier, level string, missionID primitive.ObjectID) error {
	var config models.Config
	err := lc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	placeholders := map[string]string{
		"expireRewardDays": expireRewardDays,
	}

	return lc.sendFlexMessageAndLog(userID, tier, level, missionID, config.FlexMessages.MissionComplete, placeholders)
}

func (lc *LineController) SendGetRewardFlexMessage(userID, tier, level string, missionID primitive.ObjectID) error {
	var config models.Config
	err := lc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	return lc.sendFlexMessageAndLog(userID, tier, level, missionID, config.FlexMessages.GetReward, nil)
}

func (lc *LineController) SendRewardNotificationFlexMessage(userID, remainingDays, tier, level string, missionID primitive.ObjectID) error {
	var config models.Config
	err := lc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	placeholders := map[string]string{
		"remainingDays": remainingDays,
	}

	return lc.sendFlexMessageAndLog(userID, tier, level, missionID, config.FlexMessages.RewardNotification, placeholders)
}
