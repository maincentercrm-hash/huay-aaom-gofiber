package main

import (
	"context"
	"log"
	"os"

	"go-server/config"
	"go-server/controllers"
	"go-server/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	app := fiber.New()
	app.Use(cors.New())

	// เชื่อมต่อกับ MongoDB โดยใช้ฟังก์ชัน ConnectDB จาก package config
	client, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// สร้าง context สำหรับการปิดการเชื่อมต่อ
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ทำการ defer disconnect โดยใช้ context
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from database: %v", err)
		}
	}()

	// เลือกฐานข้อมูลและ collection
	db := client.Database(os.Getenv("DB_NAME"))
	eventCollection := db.Collection("tbl_events")
	missionCollection := db.Collection("tbl_mission")
	configCollection := db.Collection("tbl_config")
	messageCollection := db.Collection("tbl_logs_message")

	// สร้าง LineController
	lineController, err := controllers.NewLineController(configCollection, messageCollection)
	if err != nil {
		log.Fatal("Failed to create LINE controller:", err)
	}

	expirationEventController := controllers.NewExpirationEventController(
		eventCollection,
		missionCollection,
		configCollection,
		lineController,
	)

	// Start background process for processing expiration events
	go expirationEventController.ProcessEvents()

	// ตั้งค่า routes
	routes.SetupGenericRoutes(app, db.Collection("tbl_users"), []string{"email", "createDate", "role", "status", "_id"}, []string{"created_at"})
	routes.SetupGenericRoutes(app, db.Collection("tbl_mission"), []string{"user_id", "status", "created_at", "updated_at", "current_tier"}, []string{"created_at"})
	routes.SetupGenericRoutes(app, db.Collection("tbl_logs_message"), []string{"user_id", "status", "sent_at"}, []string{"status", "sent_at"})
	routes.SetupAdminRoutes(app, db)
	routes.SetupConfigRoutes(app, db)
	routes.SetupClientRoutes(app, db)
	routes.SetupMissionRoutes(app, db)
	routes.SetupUserBetRoutes(app, db)
	routes.SetupDashboardRoutes(app, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Fatal(app.Listen(":" + port))
}
