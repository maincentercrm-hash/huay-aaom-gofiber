package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DashboardController struct {
	missionCollection *mongo.Collection
	clientCollection  *mongo.Collection
	userBetCollection *mongo.Collection
	logCollection     *mongo.Collection
}

func NewDashboardController(missionCollection, clientCollection, userBetCollection, logCollection *mongo.Collection) *DashboardController {
	return &DashboardController{
		missionCollection: missionCollection,
		clientCollection:  clientCollection,
		userBetCollection: userBetCollection,
		logCollection:     logCollection,
	}
}

func (dc *DashboardController) GetDashboardData(c *fiber.Ctx) error {
	ctx := context.Background()

	// 1. Get total missions
	totalMissions, err := dc.missionCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get total missions",
		})
	}

	// 2. Get tier statistics
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{}}},
		bson.D{{Key: "$unwind", Value: "$tiers"}},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$tiers.name"},
				{Key: "total_users", Value: bson.M{"$sum": 1}},
				{Key: "level_stats", Value: bson.M{
					"$push": "$tiers.current_level",
				}},
				{Key: "status_stats", Value: bson.M{
					"$push": "$tiers.status",
				}},
				{Key: "user_details", Value: bson.M{
					"$push": bson.D{
						{Key: "phone_number", Value: "$phone_number"},
						{Key: "current_tier", Value: "$current_tier"},
						{Key: "current_level", Value: "$tiers.current_level"},
						{Key: "status", Value: "$tiers.status"},
						{Key: "target", Value: "$tiers.target"},
						{Key: "last_updated", Value: "$updated_at"},
					},
				}},
			}},
		},
	}

	cursor, err := dc.missionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get tier statistics",
		})
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode results",
		})
	}

	// 3. Process results
	type TierStat struct {
		Name        string      `json:"name"`
		TotalUsers  int         `json:"total_users"`
		Completed   int         `json:"completed_users"`
		Processing  int         `json:"processing_users"`
		Failed      int         `json:"failed_users"`
		LevelCounts map[int]int `json:"level_counts"`
		UserDetails []struct {
			PhoneNumber  string    `json:"phone_number"`
			CurrentTier  int       `json:"current_tier"`
			CurrentLevel int       `json:"current_level"`
			Status       string    `json:"status"`
			Target       int       `json:"target"`
			LastUpdated  time.Time `json:"last_updated"`
		} `json:"user_details"`
	}

	tierStats := make([]TierStat, 0)
	for _, result := range results {
		tierStat := TierStat{
			Name:        result["_id"].(string),
			TotalUsers:  int(result["total_users"].(int32)),
			LevelCounts: make(map[int]int),
		}

		// Process status stats
		if statusArray, ok := result["status_stats"].(primitive.A); ok {
			for _, status := range statusArray {
				switch status.(string) {
				case "completed":
					tierStat.Completed++
				case "processing":
					tierStat.Processing++
				case "failed":
					tierStat.Failed++
				}
			}
		}

		// Process level stats
		if levelArray, ok := result["level_stats"].(primitive.A); ok {
			for _, level := range levelArray {
				if levelInt, ok := level.(int32); ok {
					tierStat.LevelCounts[int(levelInt)]++
				}
			}
		}

		// Process user details
		if details, ok := result["user_details"].(primitive.A); ok {
			for _, detail := range details {
				if detailMap, ok := detail.(primitive.M); ok {
					userDetail := struct {
						PhoneNumber  string    `json:"phone_number"`
						CurrentTier  int       `json:"current_tier"`
						CurrentLevel int       `json:"current_level"`
						Status       string    `json:"status"`
						Target       int       `json:"target"`
						LastUpdated  time.Time `json:"last_updated"`
					}{
						PhoneNumber:  detailMap["phone_number"].(string),
						CurrentTier:  int(detailMap["current_tier"].(int32)),
						CurrentLevel: int(detailMap["current_level"].(int32)),
						Status:       detailMap["status"].(string),
						Target:       int(detailMap["target"].(int32)),
						LastUpdated:  detailMap["last_updated"].(primitive.DateTime).Time(),
					}
					tierStat.UserDetails = append(tierStat.UserDetails, userDetail)
				}
			}
		}

		tierStats = append(tierStats, tierStat)
	}

	return c.JSON(fiber.Map{
		"total_missions": totalMissions,
		"tier_stats":     tierStats,
	})
}

// GetStatsData - KPI Cards สำหรับ Dashboard
func (dc *DashboardController) GetStatsData(c *fiber.Ctx) error {
	ctx := context.Background()

	// 1. ผู้ใช้ทั้งหมด (Total Clients)
	totalClients, err := dc.clientCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get total clients",
		})
	}

	// 2. มิชชันกำลังทำ (Active Missions)
	activeMissions, err := dc.missionCollection.CountDocuments(ctx, bson.M{
		"status": "processing",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get active missions",
		})
	}

	// 3. สำเร็จแล้ว (Completed Missions)
	completedMissions, err := dc.missionCollection.CountDocuments(ctx, bson.M{
		"status": "completed",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get completed missions",
		})
	}

	// 4. รางวัลรอแจก (Pending Rewards) - จาก logs ที่ status = "pending"
	pendingRewards, err := dc.logCollection.CountDocuments(ctx, bson.M{
		"status": "pending",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get pending rewards",
		})
	}

	// คำนวณ trends (เปรียบเทียบกับ 7 วันที่แล้ว)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	// Trend สำหรับ clients ใหม่
	newClientsLast7Days, _ := dc.clientCollection.CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": primitive.NewDateTimeFromTime(sevenDaysAgo)},
	})

	// Trend สำหรับ missions ที่เริ่มใหม่
	newMissionsLast7Days, _ := dc.missionCollection.CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": sevenDaysAgo},
	})

	// Trend สำหรับ missions ที่สำเร็จ
	completedLast7Days, _ := dc.missionCollection.CountDocuments(ctx, bson.M{
		"status":     "completed",
		"updated_at": bson.M{"$gte": sevenDaysAgo},
	})

	// Trend สำหรับ pending rewards (เมื่อวาน)
	yesterday := time.Now().AddDate(0, 0, -1)
	pendingYesterday, _ := dc.logCollection.CountDocuments(ctx, bson.M{
		"status":     "pending",
		"created_at": bson.M{"$gte": primitive.NewDateTimeFromTime(yesterday)},
	})

	// สร้าง response data ตาม format ของ mockData
	statsData := []fiber.Map{
		{
			"title": "ผู้ใช้ทั้งหมด",
			"value": totalClients,
			"icon":  "tabler-users",
			"color": "primary",
			"trend": fiber.Map{
				"value": newClientsLast7Days,
				"label": "จากสัปดาห์ที่แล้ว",
			},
		},
		{
			"title": "มิชชันกำลังทำ",
			"value": activeMissions,
			"icon":  "tabler-rocket",
			"color": "info",
			"trend": fiber.Map{
				"value": newMissionsLast7Days,
				"label": "จากสัปดาห์ที่แล้ว",
			},
		},
		{
			"title": "สำเร็จแล้ว",
			"value": completedMissions,
			"icon":  "tabler-circle-check",
			"color": "success",
			"trend": fiber.Map{
				"value": completedLast7Days,
				"label": "จากสัปดาห์ที่แล้ว",
			},
		},
		{
			"title": "รางวัลรอแจก",
			"value": pendingRewards,
			"icon":  "tabler-gift",
			"color": "warning",
			"trend": fiber.Map{
				"value": pendingYesterday,
				"label": "จากเมื่อวาน",
			},
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    statsData,
	})
}

// GetTierPerformanceData - ข้อมูลประสิทธิภาพแต่ละ Tier
func (dc *DashboardController) GetTierPerformanceData(c *fiber.Ctx) error {
	ctx := context.Background()

	// Aggregation pipeline เพื่อดึงข้อมูล tier performance
	pipeline := mongo.Pipeline{
		// Unwind tiers เพื่อประมวลผลแต่ละ tier
		bson.D{{Key: "$unwind", Value: "$tiers"}},

		// Group by tier name และคำนวณสถิติ
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$tiers.name"},
				{Key: "totalUsers", Value: bson.M{"$sum": 1}},
				{Key: "completedUsers", Value: bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$tiers.status", "completed"}},
							1, 0,
						},
					},
				}},
				{Key: "processingUsers", Value: bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$tiers.status", "processing"}},
							1, 0,
						},
					},
				}},
				{Key: "failedUsers", Value: bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$tiers.status", "failed"}},
							1, 0,
						},
					},
				}},
				// รวบรวมข้อมูล active users (processing users)
				{Key: "activeUsers", Value: bson.M{
					"$push": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$tiers.status", "processing"}},
							bson.D{
								{Key: "userId", Value: "$user_id"},
								{Key: "phoneNumber", Value: "$phone_number"},
								{Key: "currentLevel", Value: "$tiers.current_level"},
								{Key: "status", Value: "$tiers.status"},
								{Key: "updatedAt", Value: "$updated_at"},
							},
							"$$REMOVE",
						},
					},
				}},
			}},
		},

		// Sort by tier name
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := dc.missionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get tier performance data",
		})
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode tier performance results",
		})
	}

	// ดึงข้อมูล client details สำหรับ active users
	tierPerformanceData := make([]fiber.Map, 0)

	tierColors := map[string]string{
		"TIER 1": "#22c55e", // green
		"TIER 2": "#3b82f6", // blue
		"TIER 3": "#a855f7", // purple
	}

	for i, result := range results {
		tierName := result["_id"].(string)
		totalUsers := int(result["totalUsers"].(int32))
		completedUsers := int(result["completedUsers"].(int32))
		processingUsers := int(result["processingUsers"].(int32))
		failedUsers := int(result["failedUsers"].(int32))

		// คำนวณ success rate
		successRate := 0
		if totalUsers > 0 {
			successRate = int((float64(completedUsers) / float64(totalUsers)) * 100)
		}

		// ประมวลผล active users
		activeUsers := make([]fiber.Map, 0)
		if activeUsersArray, ok := result["activeUsers"].(primitive.A); ok {
			for j, userInterface := range activeUsersArray {
				if j >= 4 { // จำกัด 4 รายการตาม mockData
					break
				}
				if userMap, ok := userInterface.(primitive.M); ok {
					// ดึงข้อมูล client details
					var client bson.M
					err := dc.clientCollection.FindOne(ctx, bson.M{
						"user_id": userMap["userId"],
					}).Decode(&client)

					displayName := "Unknown User"
					pictureUrl := "https://i.pravatar.cc/150?img=" + fmt.Sprintf("%d", (j%50)+1)

					if err == nil {
						if name, ok := client["display_name"].(string); ok && name != "" {
							displayName = name
						}
						if pic, ok := client["picture_url"].(string); ok && pic != "" {
							pictureUrl = pic
						}
					}

					// คำนวณเวลาที่ผ่านมา (mock)
					timeLabels := []string{"5 นาทีที่แล้ว", "12 นาทีที่แล้ว", "25 นาทีที่แล้ว", "1 ชั่วโมงที่แล้ว"}
					updatedLabel := timeLabels[j%len(timeLabels)]

					activeUser := fiber.Map{
						"id":           fmt.Sprintf("%d", j+1),
						"userId":       userMap["userId"],
						"displayName":  displayName,
						"pictureUrl":   pictureUrl,
						"phoneNumber":  userMap["phoneNumber"],
						"currentLevel": userMap["currentLevel"],
						"status":       userMap["status"],
						"updatedAt":    updatedLabel,
					}
					activeUsers = append(activeUsers, activeUser)
				}
			}
		}

		// กำหนดสี tier
		tierColor := tierColors[tierName]
		if tierColor == "" {
			tierColor = "#6b7280" // default gray
		}

		tierData := fiber.Map{
			"tier":            i + 1,
			"name":            tierName,
			"totalUsers":      totalUsers,
			"completedUsers":  completedUsers,
			"processingUsers": processingUsers,
			"failedUsers":     failedUsers,
			"successRate":     successRate,
			"color":           tierColor,
			"activeUsers":     activeUsers,
		}

		tierPerformanceData = append(tierPerformanceData, tierData)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tierPerformanceData,
	})
}

// GetUrgentAlertsData - ข้อมูลการแจ้งเตือนเร่งด่วน
func (dc *DashboardController) GetUrgentAlertsData(c *fiber.Ctx) error {
	ctx := context.Background()

	// 1. รางวัลใกล้หมดอายุ (Reward Expiring) - จาก expiration_events
	now := time.Now()
	next24Hours := now.Add(24 * time.Hour)

	// ตรวจสอบจาก ExpirationEvent collection
	var expirationCollection *mongo.Collection
	// ใช้ collection tbl_expiration_events ถ้ามี หรือคำนวณจาก missions
	expirationCollection = dc.missionCollection.Database().Collection("tbl_expiration_events")

	rewardExpiringCount, err := expirationCollection.CountDocuments(ctx, bson.M{
		"type":        "reward_expiration",
		"expire_time": bson.M{"$lte": primitive.NewDateTimeFromTime(next24Hours)},
		"status":      "pending",
	})
	if err != nil {
		// ถ้าไม่มี expiration_events collection ให้คำนวณจาก missions
		rewardExpiringCount, _ = dc.missionCollection.CountDocuments(ctx, bson.M{
			"tiers.expire_reward": bson.M{
				"$lte": next24Hours,
				"$gte": now,
			},
			"status": "completed",
		})
	}

	// 2. รอการอนุมัติ (Approval Pending) - จาก logs ที่ status = "pending"
	approvalPendingCount, err := dc.logCollection.CountDocuments(ctx, bson.M{
		"status": "pending",
	})
	if err != nil {
		approvalPendingCount = 0
	}

	// 3. Level หมดอายุวันนี้ (Level Expiring Today)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	levelExpiringCount, err := expirationCollection.CountDocuments(ctx, bson.M{
		"type":        "level_expiration",
		"expire_time": bson.M{"$gte": primitive.NewDateTimeFromTime(startOfDay), "$lt": primitive.NewDateTimeFromTime(endOfDay)},
		"status":      "pending",
	})
	if err != nil {
		// ถ้าไม่มี expiration_events collection ให้คำนวณจาก missions
		pipeline := mongo.Pipeline{
			bson.D{{Key: "$unwind", Value: "$tiers"}},
			bson.D{{Key: "$unwind", Value: "$tiers.levels"}},
			bson.D{{Key: "$match", Value: bson.M{
				"tiers.levels.expire_date": bson.M{
					"$gte": startOfDay,
					"$lt":  endOfDay,
				},
				"tiers.levels.status": bson.M{"$ne": "completed"},
			}}},
			bson.D{{Key: "$count", Value: "total"}},
		}

		cursor, err := dc.missionCollection.Aggregate(ctx, pipeline)
		if err == nil {
			var result []bson.M
			if err := cursor.All(ctx, &result); err == nil && len(result) > 0 {
				if total, ok := result[0]["total"].(int32); ok {
					levelExpiringCount = int64(total)
				}
			}
			cursor.Close(ctx)
		}
	}

	// สร้าง alerts data ตาม format ของ mockData
	alertsData := []fiber.Map{
		{
			"id":          "1",
			"type":        "reward_expiring",
			"title":       "รางวัลใกล้หมดอายุ",
			"description": "มีรางวัลที่จะหมดอายุภายใน 24 ชั่วโมง",
			"count":       rewardExpiringCount,
			"severity":    "error",
			"icon":        "tabler-alarm",
			"actionLabel": "ดูรายละเอียด",
		},
		{
			"id":          "2",
			"type":        "approval_pending",
			"title":       "รอการอนุมัติ",
			"description": "มีรางวัลรอการอนุมัติจากระบบ",
			"count":       approvalPendingCount,
			"severity":    "warning",
			"icon":        "tabler-hourglass",
			"actionLabel": "ดำเนินการ",
		},
		{
			"id":          "3",
			"type":        "level_expiring",
			"title":       "Level หมดอายุวันนี้",
			"description": "มี Level ที่จะหมดอายุวันนี้",
			"count":       levelExpiringCount,
			"severity":    "info",
			"icon":        "tabler-calendar-event",
			"actionLabel": "ดูรายชื่อ",
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alertsData,
	})
}

// GetRecentActivitiesData - ข้อมูลกิจกรรมล่าสุด
func (dc *DashboardController) GetRecentActivitiesData(c *fiber.Ctx) error {
	ctx := context.Background()

	// ดึง limit จาก query parameter (default = 10)
	limit := c.QueryInt("limit", 10)

	// Aggregation pipeline เพื่อดึง recent activities จาก logs
	pipeline := mongo.Pipeline{
		// เรียงตาม created_at ล่าสุด
		bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},

		// จำกัดจำนวน
		bson.D{{Key: "$limit", Value: limit}},
	}

	cursor, err := dc.logCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get recent activities",
		})
	}
	defer cursor.Close(ctx)

	var logs []bson.M
	if err := cursor.All(ctx, &logs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode recent activities",
		})
	}

	// แปลง logs เป็น activities format
	activities := make([]fiber.Map, 0)

	for i, log := range logs {
		// ดึงข้อมูล user
		var client bson.M
		var userInfo fiber.Map

		userId, _ := log["user_id"].(string)
		if userId != "" {
			err := dc.clientCollection.FindOne(ctx, bson.M{
				"user_id": userId,
			}).Decode(&client)

			if err == nil {
				displayName := "Unknown User"
				pictureUrl := "https://i.pravatar.cc/150?img=" + fmt.Sprintf("%d", (i%50)+1)

				if name, ok := client["display_name"].(string); ok && name != "" {
					displayName = name
				}
				if pic, ok := client["picture_url"].(string); ok && pic != "" {
					pictureUrl = pic
				}

				userInfo = fiber.Map{
					"userId":      userId,
					"displayName": displayName,
					"pictureUrl":  pictureUrl,
				}
			}
		}

		// กำหนด activity type, title และ description ตามโครงสร้างข้อมูลจริงของ tbl_logs
		activityType := "mission_reward"
		title := "ได้รับรางวัล"
		description := "รอการอนุมัติ"
		metadata := fiber.Map{}

		// ดึงข้อมูลจาก log
		status, _ := log["status"].(string)
		missionDetail, _ := log["mission_detail"].(string)

		// กำหนด activity type และ title ตาม status
		switch status {
		case "approve":
			activityType = "user_claimed_reward"
			title = "ผู้ใช้รับรางวัล"
			description = "รับรางวัลจากระบบสำเร็จ"

		case "pending":
			activityType = "reward_pending"
			title = "รางวัลรออนุมัติ"
			description = "รอการอนุมัติจากระบบ"

		case "rejected":
			activityType = "reward_rejected"
			title = "รางวัลถูกปฏิเสธ"
			description = "ไม่ผ่านการอนุมัติ"

		case "completed":
			activityType = "user_completed_mission"
			title = "ทำภารกิจสำเร็จ"
			description = "ทำภารกิจสำเร็จและได้รับรางวัล"
		}

		// ใช้ mission_detail ถ้ามี (แสดงรายละเอียดเต็ม)
		if missionDetail != "" {
			metadata["missionDetail"] = missionDetail
		}

		// เพิ่ม reward amount ถ้ามี
		if reward, ok := log["reward"].(float64); ok {
			metadata["rewardAmount"] = reward
		} else if reward, ok := log["reward"].(int32); ok {
			metadata["rewardAmount"] = float64(reward)
		} else if reward, ok := log["reward"].(int64); ok {
			metadata["rewardAmount"] = float64(reward)
		}

		// เพิ่ม mission_id ถ้ามี
		if missionId, ok := log["mission_id"].(string); ok {
			metadata["missionId"] = missionId
		}

		// เพิ่ม callback_time ถ้ามี (สำหรับ approved rewards)
		if callbackTime, ok := log["callback_time"].(primitive.DateTime); ok {
			metadata["callbackTime"] = callbackTime.Time().Format("02/01/2006 15:04")
		}

		// คำนวณ timestamp จาก created_at
		timestamp := "เมื่อสักครู่"
		if createdAt, ok := log["created_at"].(primitive.DateTime); ok {
			activityTime := createdAt.Time()
			duration := time.Since(activityTime)

			if duration.Hours() >= 24 {
				days := int(duration.Hours() / 24)
				timestamp = fmt.Sprintf("%d วันที่แล้ว", days)
			} else if duration.Hours() >= 1 {
				hours := int(duration.Hours())
				timestamp = fmt.Sprintf("%d ชั่วโมงที่แล้ว", hours)
			} else if duration.Minutes() >= 1 {
				minutes := int(duration.Minutes())
				timestamp = fmt.Sprintf("%d นาทีที่แล้ว", minutes)
			} else {
				seconds := int(duration.Seconds())
				timestamp = fmt.Sprintf("%d วินาทีที่แล้ว", seconds)
			}
		}

		activity := fiber.Map{
			"id":          log["_id"],
			"type":        activityType,
			"title":       title,
			"description": description,
			"timestamp":   timestamp,
			"status":      status,
		}

		// เพิ่ม user info ถ้ามี
		if userInfo != nil && len(userInfo) > 0 {
			activity["user"] = userInfo
		}

		// เพิ่ม metadata ถ้ามี
		if len(metadata) > 0 {
			activity["metadata"] = metadata
		}

		activities = append(activities, activity)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    activities,
	})
}

// GetPendingRewards - สรุปภาพรวมรางวัล (pending/approved/rejected) และรายการที่รออนุมัติพร้อมข้อมูล client
func (dc *DashboardController) GetPendingRewards(c *fiber.Ctx) error {
	ctx := context.Background()

	// ดึง limit จาก query parameter (default = 50)
	limit := c.QueryInt("limit", 50)

	// Aggregation pipeline เพื่อดึงรายการ pending rewards พร้อม client info
	pipeline := mongo.Pipeline{
		// Filter เฉพาะ status = "pending"
		bson.D{{Key: "$match", Value: bson.M{"status": "pending"}}},

		// เรียงตาม created_at ล่าสุด
		bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},

		// จำกัดจำนวน
		bson.D{{Key: "$limit", Value: limit}},
	}

	cursor, err := dc.logCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get pending rewards",
		})
	}
	defer cursor.Close(ctx)

	var logs []bson.M
	if err := cursor.All(ctx, &logs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode pending rewards",
		})
	}

	// นับจำนวนรายการ pending ทั้งหมด
	totalPending, err := dc.logCollection.CountDocuments(ctx, bson.M{"status": "pending"})
	if err != nil {
		totalPending = 0
	}

	// นับจำนวนรายการ approved และ rejected
	totalApproved, _ := dc.logCollection.CountDocuments(ctx, bson.M{"status": "approve"})
	totalRejected, _ := dc.logCollection.CountDocuments(ctx, bson.M{"status": "reject"})

	// คำนวณยอดเงินที่จ่ายไปแล้ว (approved)
	approvedPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"status": "approve"}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "totalAmount", Value: bson.M{"$sum": "$reward"}},
		}}},
	}

	approvedCursor, err := dc.logCollection.Aggregate(ctx, approvedPipeline)
	if err == nil {
		defer approvedCursor.Close(ctx)
	}

	var totalPaidAmount float64
	if err == nil && approvedCursor.Next(ctx) {
		var result bson.M
		if err := approvedCursor.Decode(&result); err == nil {
			if amount, ok := result["totalAmount"].(float64); ok {
				totalPaidAmount = amount
			} else if amount, ok := result["totalAmount"].(int32); ok {
				totalPaidAmount = float64(amount)
			} else if amount, ok := result["totalAmount"].(int64); ok {
				totalPaidAmount = float64(amount)
			}
		}
	}

	// สร้าง pending rewards list พร้อม client info
	pendingRewards := make([]fiber.Map, 0)

	for i, log := range logs {
		// ดึงข้อมูล client
		var client bson.M
		userId, _ := log["user_id"].(string)

		var displayName, phoneNumber, pictureUrl string
		displayName = "Unknown User"
		phoneNumber = "-"
		pictureUrl = "https://i.pravatar.cc/150?img=" + fmt.Sprintf("%d", (i%50)+1)

		if userId != "" {
			err := dc.clientCollection.FindOne(ctx, bson.M{
				"user_id": userId,
			}).Decode(&client)

			if err == nil {
				if name, ok := client["display_name"].(string); ok && name != "" {
					displayName = name
				}
				if phone, ok := client["phone_number"].(string); ok && phone != "" {
					phoneNumber = phone
				}
				if pic, ok := client["picture_url"].(string); ok && pic != "" {
					pictureUrl = pic
				}
			}
		}

		// ดึงข้อมูล mission
		missionId, _ := log["mission_id"].(string)
		missionDetail, _ := log["mission_detail"].(string)

		// ดึงข้อมูล reward amount
		var rewardAmount float64
		if reward, ok := log["reward"].(float64); ok {
			rewardAmount = reward
		} else if reward, ok := log["reward"].(int32); ok {
			rewardAmount = float64(reward)
		} else if reward, ok := log["reward"].(int64); ok {
			rewardAmount = float64(reward)
		}

		// คำนวณเวลาที่รอ (waiting time)
		var waitingTime string
		var createdAtStr string
		if createdAt, ok := log["created_at"].(primitive.DateTime); ok {
			createdTime := createdAt.Time()
			duration := time.Since(createdTime)

			if duration.Hours() >= 24 {
				days := int(duration.Hours() / 24)
				waitingTime = fmt.Sprintf("%d วัน", days)
			} else if duration.Hours() >= 1 {
				hours := int(duration.Hours())
				waitingTime = fmt.Sprintf("%d ชั่วโมง", hours)
			} else if duration.Minutes() >= 1 {
				minutes := int(duration.Minutes())
				waitingTime = fmt.Sprintf("%d นาที", minutes)
			} else {
				seconds := int(duration.Seconds())
				waitingTime = fmt.Sprintf("%d วินาที", seconds)
			}

			// Format created_at เป็น Thai format
			thailandLoc, _ := time.LoadLocation("Asia/Bangkok")
			createdTimeTH := createdTime.In(thailandLoc)
			createdAtStr = createdTimeTH.Format("02/01/2006 15:04")
		}

		pendingReward := fiber.Map{
			"logId":         log["_id"],
			"userId":        userId,
			"displayName":   displayName,
			"phoneNumber":   phoneNumber,
			"pictureUrl":    pictureUrl,
			"missionId":     missionId,
			"missionDetail": missionDetail,
			"rewardAmount":  rewardAmount,
			"createdAt":     createdAtStr,
			"waitingTime":   waitingTime,
			"status":        "pending",
		}

		pendingRewards = append(pendingRewards, pendingReward)
	}

	// สร้าง summary object
	summary := fiber.Map{
		"pending":   totalPending,
		"approved":  totalApproved,
		"rejected":  totalRejected,
		"totalPaid": totalPaidAmount,
		"total":     totalPending + totalApproved + totalRejected,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"summary": summary,
		"data":    pendingRewards,
	})
}
