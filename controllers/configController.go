package controllers

import (
	"context"
	"fmt"
	"go-server/models"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

type ConfigController struct {
	Collection *mongo.Collection
}

func NewConfigController(collection *mongo.Collection) *ConfigController {
	return &ConfigController{
		Collection: collection,
	}
}

func (cc *ConfigController) GetConfig(c *fiber.Ctx) error {
	var config models.Config
	err := cc.Collection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Config not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch config"})
	}
	return c.JSON(config)
}

func (cc *ConfigController) SaveConfig(c *fiber.Ctx) error {
	var config models.Config
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{}
	update := bson.M{"$set": config}

	var updatedConfig models.Config
	err := cc.Collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedConfig)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save config"})
	}

	return c.JSON(updatedConfig)
}

// New function to update only tier settings
func (cc *ConfigController) UpdateTierSettings(c *fiber.Ctx) error {
	var tierSettings struct {
		Tiers []models.TierDetail `json:"tiers"`
	}
	if err := c.BodyParser(&tierSettings); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// เพิ่ม logging
	fmt.Printf("Received tiers: %+v\n", tierSettings.Tiers)

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{}
	update := bson.M{"$set": bson.M{"tiers": tierSettings.Tiers}}

	var updatedConfig models.Config
	err := cc.Collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedConfig)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update tier settings"})
	}

	// เพิ่ม logging
	//fmt.Printf("Updated config: %+v\n", updatedConfig)

	return c.JSON(updatedConfig)
}

func (cc *ConfigController) UpdateFlexMessageSettings(c *fiber.Ctx) error {
	var flexMessagesUpdate struct {
		FlexMessages models.FlexMessages `json:"flexMessages"`
	}

	if err := c.BodyParser(&flexMessagesUpdate); err != nil {
		log.Printf("Error parsing flex messages: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Printf("Received flex messages update:")
	log.Printf("Followup: %+v", flexMessagesUpdate.FlexMessages.Followup)
	log.Printf("Mission Success: %+v", flexMessagesUpdate.FlexMessages.MissionSuccess)
	log.Printf("Mission Failed: %+v", flexMessagesUpdate.FlexMessages.MissionFailed)
	log.Printf("Mission Complete: %+v", flexMessagesUpdate.FlexMessages.MissionComplete)
	log.Printf("Get Reward: %+v", flexMessagesUpdate.FlexMessages.GetReward)
	log.Printf("Reward Notification: %+v", flexMessagesUpdate.FlexMessages.RewardNotification)

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{}
	update := bson.M{"$set": bson.M{"flex_messages": flexMessagesUpdate.FlexMessages}}

	var updatedConfig models.Config
	err := cc.Collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedConfig)
	if err != nil {
		log.Printf("Error updating flex messages: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update flex message settings"})
	}

	log.Printf("Flex messages updated successfully")
	return c.JSON(updatedConfig)
}
func (cc *ConfigController) UploadImage(c *fiber.Ctx) error {
	log.Println("Starting image upload...")

	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No file uploaded"})
	}

	uploadType := c.FormValue("type")
	if uploadType == "" {
		log.Println("Upload type is missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Upload type is required"})
	}

	log.Printf("Uploading image for type: %s", uploadType)

	var config models.Config
	err = cc.Collection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		log.Printf("Error fetching config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch config"})
	}

	bucketName := config.FirebaseConfig.BucketName
	if bucketName == "" {
		log.Println("Bucket name is not set in config")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Bucket name is not configured"})
	}

	opt := option.WithCredentialsJSON([]byte(config.FirebaseConfig.Credential))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Printf("Error initializing Firebase app: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to initialize Firebase app"})
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		log.Printf("Error getting storage client: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create storage client"})
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log.Printf("Error getting bucket %s: %v", bucketName, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get bucket"})
	}

	var filename string
	if strings.HasPrefix(uploadType, "siteTemplate") {
		filename = fmt.Sprintf("site_template/%s/%s", strings.TrimPrefix(uploadType, "siteTemplate."), filepath.Base(file.Filename))
	} else {
		filename = fmt.Sprintf("flex_messages/%s/%s", uploadType, filepath.Base(file.Filename))
	}
	log.Printf("Uploading to filename: %s", filename)

	obj := bucket.Object(filename)
	writer := obj.NewWriter(context.Background())

	// ตรวจสอบและตั้งค่า Content-Type
	src, err := file.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file"})
	}

	contentType := http.DetectContentType(buffer)

	// ตรวจสอบว่าเป็น SVG หรือไม่
	if contentType == "text/xml; charset=utf-8" || contentType == "text/plain; charset=utf-8" {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == ".svg" {
			contentType = "image/svg+xml"
		}
	}

	writer.ContentType = contentType
	log.Printf("Content-Type set to: %s", contentType)

	// Reset file pointer
	src.Seek(0, 0)

	log.Println("Copying file to storage...")
	if _, err = io.Copy(writer, src); err != nil {
		log.Printf("Error copying file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload file"})
	}

	log.Println("Closing writer...")
	if err := writer.Close(); err != nil {
		log.Printf("Error closing writer: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to close writer"})
	}

	log.Println("Making file public...")
	if err := obj.ACL().Set(context.Background(), storage.AllUsers, storage.RoleReader); err != nil {
		log.Printf("Error making file public: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to make file public"})
	}

	log.Println("Getting public URL...")
	attrs, err := obj.Attrs(context.Background())
	if err != nil {
		log.Printf("Error getting file attributes: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get file attributes"})
	}

	log.Printf("Image uploaded successfully. URL: %s", attrs.MediaLink)
	return c.JSON(fiber.Map{"imageUrl": attrs.MediaLink})
}

func (cc *ConfigController) UpdateSiteTemplateConfig(c *fiber.Ctx) error {
	var siteTemplateUpdate struct {
		SiteTemplate models.SiteTemplateConfig `json:"siteTemplate"`
	}

	if err := c.BodyParser(&siteTemplateUpdate); err != nil {
		log.Printf("Error parsing site template config: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Printf("Received site template update: %+v", siteTemplateUpdate.SiteTemplate)

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{}
	update := bson.M{"$set": bson.M{"site_template": siteTemplateUpdate.SiteTemplate}}

	var updatedConfig models.Config
	err := cc.Collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedConfig)
	if err != nil {
		log.Printf("Error updating site template config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update site template config"})
	}

	log.Printf("Site template config updated successfully")
	return c.JSON(updatedConfig)
}
