package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"go-server/models"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientController struct {
	collection       *mongo.Collection
	configCollection *mongo.Collection
}

func NewClientController(collection, configCollection *mongo.Collection) *ClientController {
	return &ClientController{
		collection:       collection,
		configCollection: configCollection,
	}
}

func (cc *ClientController) GetAllClients(c *fiber.Ctx) error {
	log.Println("GetAllClients: Starting")

	// รับ query parameters สำหรับ pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	skip := (page - 1) * limit

	// สร้าง filter สำหรับการค้นหา
	filter := bson.M{}
	if search != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"display_name": bson.M{"$regex": search, "$options": "i"}},
				{"user_id": bson.M{"$regex": search, "$options": "i"}},
				{"phone_number": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

	// ตัวเลือกสำหรับการ query
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{"updated_at", -1}}) // เรียงตามวันที่อัปเดตล่าสุด

	// ดึงข้อมูล clients
	cursor, err := cc.collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		log.Printf("GetAllClients: Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer cursor.Close(context.Background())

	var clients []models.Client
	if err = cursor.All(context.Background(), &clients); err != nil {
		log.Printf("GetAllClients: Error decoding clients: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error decoding clients"})
	}

	// นับจำนวนรวมทั้งหมด
	totalCount, err := cc.collection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Printf("GetAllClients: Error counting documents: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error counting documents"})
	}

	totalPages := (totalCount + int64(limit) - 1) / int64(limit)

	log.Printf("GetAllClients: Found %d clients (page %d of %d)", len(clients), page, totalPages)

	return c.JSON(fiber.Map{
		"clients": clients,
		"pagination": fiber.Map{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  totalCount,
			"limit":        limit,
			"has_next":     page < int(totalPages),
			"has_prev":     page > 1,
		},
	})
}

func (cc *ClientController) GetClientByUserId(c *fiber.Ctx) error {
	log.Println("GetClientByUserId: Starting")
	idParam := c.Params("userId")
	log.Printf("GetClientByUserId: Getting client for ID: %s", idParam)

	// สร้าง filter ที่รองรับทั้ง MongoDB ObjectID และ LINE user_id
	var filter bson.M

	// ลองแปลงเป็น ObjectID ก่อน
	if objectID, err := primitive.ObjectIDFromHex(idParam); err == nil {
		// ถ้าเป็น valid ObjectID ให้ค้นหาด้วย _id
		filter = bson.M{"_id": objectID}
		log.Printf("GetClientByUserId: Using MongoDB ObjectID filter: %s", idParam)
	} else {
		// ถ้าไม่ใช่ ObjectID ให้ค้นหาด้วย user_id (LINE ID)
		filter = bson.M{"user_id": idParam}
		log.Printf("GetClientByUserId: Using user_id filter: %s", idParam)
	}

	var client models.Client
	err := cc.collection.FindOne(context.Background(), filter).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("GetClientByUserId: Client not found for ID: %s", idParam)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Client not found"})
		}
		log.Printf("GetClientByUserId: Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	log.Printf("GetClientByUserId: Found client for ID: %s", idParam)
	return c.JSON(fiber.Map{
		"client": client,
	})
}

func (cc *ClientController) DeleteClient(c *fiber.Ctx) error {
	log.Println("DeleteClient: Starting")
	idParam := c.Params("userId")
	log.Printf("DeleteClient: Deleting client for ID: %s", idParam)

	// สร้าง filter ที่รองรับทั้ง MongoDB ObjectID และ LINE user_id
	var filter bson.M

	// ลองแปลงเป็น ObjectID ก่อน
	if objectID, err := primitive.ObjectIDFromHex(idParam); err == nil {
		// ถ้าเป็น valid ObjectID ให้ค้นหาด้วย _id
		filter = bson.M{"_id": objectID}
		log.Printf("DeleteClient: Using MongoDB ObjectID filter: %s", idParam)
	} else {
		// ถ้าไม่ใช่ ObjectID ให้ค้นหาด้วย user_id (LINE ID)
		filter = bson.M{"user_id": idParam}
		log.Printf("DeleteClient: Using user_id filter: %s", idParam)
	}

	// ตรวจสอบว่า client มีอยู่หรือไม่ก่อนลบ
	var client models.Client
	err := cc.collection.FindOne(context.Background(), filter).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("DeleteClient: Client not found for ID: %s", idParam)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "Client not found",
			})
		}
		log.Printf("DeleteClient: Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Database error",
		})
	}

	// ลบ client
	result, err := cc.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("DeleteClient: Error deleting client: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Error deleting client",
		})
	}

	if result.DeletedCount == 0 {
		log.Printf("DeleteClient: No client deleted for ID: %s", idParam)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Client not found",
		})
	}

	log.Printf("DeleteClient: Successfully deleted client for ID: %s", idParam)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Client deleted successfully",
		"deleted_client": fiber.Map{
			"id":           client.ID.Hex(),
			"user_id":      client.UserID,
			"display_name": client.DisplayName,
		},
	})
}

func (cc *ClientController) UpsertClient(c *fiber.Ctx) error {
	log.Println("UpsertClient: Starting")
	var client models.Client
	if err := c.BodyParser(&client); err != nil {
		log.Printf("UpsertClient: Error parsing body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	log.Printf("UpsertClient: Received client data: %+v", client)

	now := primitive.NewDateTimeFromTime(time.Now())
	filter := bson.M{"user_id": client.UserID}
	update := bson.M{
		"$set": bson.M{
			"display_name":   client.DisplayName,
			"picture_url":    client.PictureURL,
			"status_message": client.StatusMessage,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}
	opts := options.Update().SetUpsert(true)

	result, err := cc.collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		log.Printf("UpsertClient: Error updating database: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	log.Printf("UpsertClient: Database update result: %+v", result)

	// ดึงข้อมูล user ล่าสุดหลังจาก upsert
	var updatedClient models.Client
	err = cc.collection.FindOne(context.Background(), filter).Decode(&updatedClient)
	if err != nil {
		log.Printf("UpsertClient: Error fetching updated client: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error fetching updated client"})
	}

	if result.UpsertedID != nil {
		log.Println("UpsertClient: Client created")
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Client created",
			"id":      result.UpsertedID,
			"client":  updatedClient,
		})
	}

	log.Println("UpsertClient: Client updated")
	return c.JSON(fiber.Map{
		"message": "Client updated",
		"client":  updatedClient,
	})
}

func (cc *ClientController) CheckPhoneNumber(c *fiber.Ctx) error {
	log.Println("CheckPhoneNumber: Starting")
	userID := c.Params("userId")
	log.Printf("CheckPhoneNumber: Checking phone number for user ID: %s", userID)

	var client models.Client
	err := cc.collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("CheckPhoneNumber: Client not found for user ID: %s", userID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Client not found"})
		}
		log.Printf("CheckPhoneNumber: Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	hasPhoneNumber := client.PhoneNumber != ""
	log.Printf("CheckPhoneNumber: User ID %s has phone number: %v", userID, hasPhoneNumber)
	return c.JSON(fiber.Map{"hasPhoneNumber": hasPhoneNumber})
}

func (cc *ClientController) checkWithExternalAPI(phoneNumber, lineID string, config *models.Config) (bool, string, error) {
	log.Printf("checkWithExternalAPI: Checking phone number %s for LINE ID %s", phoneNumber, lineID)
	client := &http.Client{}
	data := map[string]string{
		"phone_number": phoneNumber,
		"line_id":      lineID,
		"line_at":      config.LineAt,
	}

	// Log request data ก่อน marshal
	log.Printf("checkWithExternalAPI: Request data: %+v", data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("checkWithExternalAPI: Error marshaling JSON: %v", err)
		return false, "", err
	}

	req, err := http.NewRequest("POST", config.ApiEndpoint+"/players/v1/line/sync", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("checkWithExternalAPI: Error creating request: %v", err)
		return false, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-KEY", config.ApiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("checkWithExternalAPI: Error sending request: %v", err)
		return false, "", err
	}
	defer resp.Body.Close()

	log.Printf("checkWithExternalAPI: Response status code: %d", resp.StatusCode)

	// อ่าน response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("checkWithExternalAPI: Error reading response body: %v", err)
		return false, "", err
	}

	// แสดง raw response
	log.Printf("checkWithExternalAPI: Response body: %s", string(bodyBytes))

	// สร้าง reader ใหม่จาก bytes เพื่อใช้ decode JSON
	reader := bytes.NewReader(bodyBytes)

	// ยอมรับทั้ง 200 OK และ 201 Created
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("checkWithExternalAPI: External API returned unexpected status: %d", resp.StatusCode)
		return false, "", nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		log.Printf("checkWithExternalAPI: Error decoding response: %v", err)
		return false, "", err
	}

	username, exists := result["username"].(string)
	log.Printf("checkWithExternalAPI: Username exists in response: %v", exists)
	return exists, username, nil
}

func (cc *ClientController) UpdatePhoneNumber(c *fiber.Ctx) error {
	log.Println("UpdatePhoneNumber: Starting")
	userID := c.Params("userId")
	log.Printf("UpdatePhoneNumber: Updating phone number for user ID: %s", userID)

	var updateData struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.BodyParser(&updateData); err != nil {
		log.Printf("UpdatePhoneNumber: Error parsing body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Cannot parse JSON"})
	}
	log.Printf("UpdatePhoneNumber: Received phone number: %s", updateData.PhoneNumber)

	var config models.Config
	err := cc.configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		log.Printf("UpdatePhoneNumber: Error fetching config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to fetch config"})
	}

	isValid, username, err := cc.checkWithExternalAPI(updateData.PhoneNumber, userID, &config)
	if err != nil {
		log.Printf("UpdatePhoneNumber: Error checking with external API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to check with external API"})
	}

	if !isValid {
		log.Printf("UpdatePhoneNumber: Invalid phone number or user ID for user: %s", userID)
		return c.JSON(fiber.Map{
			"success":       false,
			"line_sync_url": config.LineSyncURL,
		})
	}

	// New check: Compare phone number with username
	if updateData.PhoneNumber != username {
		log.Printf("UpdatePhoneNumber: Phone number does not match username for user: %s", userID)
		return c.JSON(fiber.Map{
			"success":         false,
			"error":           "Phone number does not match the data in the system",
			"received_number": updateData.PhoneNumber,
			"expected_number": username,
		})
	}

	update := bson.M{
		"$set": bson.M{
			"phone_number": updateData.PhoneNumber,
			"updated_at":   primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	result, err := cc.collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, update)
	if err != nil {
		log.Printf("UpdatePhoneNumber: Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Database error"})
	}

	if result.MatchedCount == 0 {
		log.Printf("UpdatePhoneNumber: Client not found for user ID: %s", userID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Client not found"})
	}

	log.Printf("UpdatePhoneNumber: Phone number updated successfully for user ID: %s", userID)
	return c.JSON(fiber.Map{"success": true, "message": "Phone number updated successfully"})
}
