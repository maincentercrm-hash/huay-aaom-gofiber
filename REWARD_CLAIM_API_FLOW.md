# Reward Claim API Flow Documentation
# เอกสารอธิบายขั้นตอนการขอรับรางวัลและการเชื่อมต่อ API ภายนอก

> **Version:** 1.0
> **Last Updated:** 2025-12-11
> **Related Files:** `controllers/missionController.go`, `controllers/rewardCallbackController.go`, `utils/getCurrentBet.go`

---

## สารบัญ (Table of Contents)

1. [ภาพรวมระบบ (System Overview)](#1-ภาพรวมระบบ-system-overview)
2. [Flow Diagram](#2-flow-diagram)
3. [ขั้นตอนที่ 1: เมื่อ User ผ่าน Tier ครบ 3 Levels](#3-ขั้นตอนที่-1-เมื่อ-user-ผ่าน-tier-ครบ-3-levels)
4. [ขั้นตอนที่ 2: User กดขอรับรางวัล (Claim Reward)](#4-ขั้นตอนที่-2-user-กดขอรับรางวัล-claim-reward)
5. [ขั้นตอนที่ 3: ส่งข้อมูลไปยัง External API](#5-ขั้นตอนที่-3-ส่งข้อมูลไปยัง-external-api)
6. [ขั้นตอนที่ 4: External API Callback กลับมา](#6-ขั้นตอนที่-4-external-api-callback-กลับมา)
7. [External API Endpoints ทั้งหมด](#7-external-api-endpoints-ทั้งหมด)
8. [Authentication](#8-authentication)
9. [Error Handling](#9-error-handling)

---

## 1. ภาพรวมระบบ (System Overview)

ระบบ Retirement Lottery มีการเชื่อมต่อกับ **External Players API** เพื่อ:
1. **ตรวจสอบเบอร์โทร** - ยืนยันว่าเบอร์โทรตรงกับ LINE ID ในระบบปลายทาง
2. **ดึงยอดเดิมพัน** - ดึงยอด bet ของ user ในช่วงเวลาที่กำหนด
3. **ส่งคำขอรับรางวัล** - ส่งข้อมูลรางวัลไปให้ระบบปลายทางอนุมัติ

### Configuration ที่เกี่ยวข้อง (จาก `tbl_config`)

```json
{
  "api_endpoint": "https://example.com/api",  // Base URL ของ External API
  "api_key": "your-api-key-here",             // API Key สำหรับ Authentication
  "line_at": "@lineofficial"                  // LINE Official Account ID
}
```

---

## 2. Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         REWARD CLAIM FLOW                                    │
└─────────────────────────────────────────────────────────────────────────────┘

                    ┌─────────────────┐
                    │   User ผ่าน     │
                    │ Tier 1 ครบ 3    │
                    │    Levels       │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  Tier Status    │
                    │  = "completed"  │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  แสดงปุ่ม        │
                    │ "Get Reward"    │
                    │  ให้ User กด    │
                    └────────┬────────┘
                             │
                             ▼
        ┌────────────────────────────────────────┐
        │       POST /api/missions/:id/claim-reward       │
        └────────────────────┬───────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
     ┌────────────┐  ┌────────────┐  ┌────────────┐
     │  สร้าง Log │  │ ส่ง LINE   │  │ส่ง Telegram│
     │  (pending) │  │ Message    │  │ แจ้ง Admin │
     └─────┬──────┘  └────────────┘  └────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────┐
│    POST {api_endpoint}/players/v1/line/rewards/claim            │
│                                                                  │
│    Request:                                                      │
│    {                                                             │
│      "log_id": "...",                                           │
│      "user_id": "LINE_USER_ID",                                 │
│      "mission_detail": "Tier 1 Complete...",                    │
│      "reward": 100.00,                                          │
│      "callback_url": "https://your-server/api/missions/reward-callback",│
│      "line_at": "@lineofficial"                                 │
│    }                                                             │
│                                                                  │
│    Headers:                                                      │
│    - Content-Type: application/json                              │
│    - api-key: {api_key}                                         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  External API   │
                    │  ประมวลผล        │
                    │  (อนุมัติ/ปฏิเสธ)  │
                    └────────┬────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│         POST {callback_url} (Callback from External)            │
│                                                                  │
│    Request:                                                      │
│    {                                                             │
│      "log_id": "...",                                           │
│      "status": "approve" | "reject"                             │
│    }                                                             │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
              ▼                             ▼
     ┌────────────────┐            ┌────────────────┐
     │ status=approve │            │ status=reject  │
     │                │            │                │
     │ - อัปเดต Log    │            │ - อัปเดต Log    │
     │ - สร้าง Tier    │            │ - Tier กลับไป   │
     │   ถัดไป         │            │  awaiting_reward│
     │ - Mission      │            │                │
     │   ดำเนินต่อ     │            │                │
     └────────────────┘            └────────────────┘
```

---

## 3. ขั้นตอนที่ 1: เมื่อ User ผ่าน Tier ครบ 3 Levels

### 3.1 เงื่อนไขการผ่าน Level

เมื่อ **Level หมดอายุ** (Expiration Event) ระบบจะตรวจสอบ:

```go
// จาก controllers/missionController.go:177
if currentBet >= float64(currentTierConfig.Target) {
    // ผ่าน Level
    currentLevel.Status = "success"
} else {
    // ไม่ผ่าน Level
    currentLevel.Status = "failed"
}
```

### 3.2 เมื่อผ่านครบ 3 Levels (Tier Completed)

```go
// จาก controllers/missionController.go:198-201
if currentTier.CurrentLevel >= currentTier.MaxLevel {
    // Level สุดท้ายของ Tier
    currentTier.Status = "completed"
    log.Printf("Tier %d completed! Reward: %d", tierIndex+1, currentTier.Reward)
}
```

**สถานะ Tier หลังผ่านครบ:**
- `Tier.Status = "completed"`
- ปุ่ม "Get Reward" จะแสดงให้ User เห็นบน Frontend

---

## 4. ขั้นตอนที่ 2: User กดขอรับรางวัล (Claim Reward)

### 4.1 API Endpoint

```
POST /api/missions/:id/claim-reward
```

### 4.2 Flow การทำงาน

**Source:** `controllers/missionController.go:242-346`

```go
func (c *MissionController) ClaimReward(ctx *fiber.Ctx) error {
    // 1. ตรวจสอบ Mission ID
    missionID, err := primitive.ObjectIDFromHex(ctx.Params("id"))

    // 2. ตรวจสอบว่า Tier มีสิทธิ์รับรางวัล
    currentTier := &mission.Tiers[mission.CurrentTier-1]
    if currentTier.Status != "completed" && currentTier.Status != "awaiting_reward" {
        return error("Current tier is not eligible for reward")
    }

    // 3. ลบ Events ที่เกี่ยวข้องกับ Reward
    c.clearRewardRelatedEvents(ctx.Context(), missionID)

    // 4. อัปเดต Mission Status เป็น "pending"
    missionCollection.UpdateOne({
        "status": "pending",
        "tiers.X.status": "pending"
    })

    // 5. ส่งแจ้งเตือน Telegram ให้ Admin
    telegramController.SendRewardClaimedMessage(...)

    // 6. ส่งแจ้งเตือน LINE ให้ User
    lineController.SendGetRewardFlexMessage(...)

    // 7. ส่งคำขอรางวัลไปยัง External API
    c.sendRewardClaimToExternalAPI(...)
}
```

### 4.3 Response

```json
{
    "message": "Reward claim sent successfully",
    "status": "pending"
}
```

---

## 5. ขั้นตอนที่ 3: ส่งข้อมูลไปยัง External API

### 5.1 API Endpoint ปลายทาง

```
POST {api_endpoint}/players/v1/line/rewards/claim
```

**Example:**
```
POST https://example.com/api/players/v1/line/rewards/claim
```

### 5.2 Request Headers

| Header | Value | Description |
|--------|-------|-------------|
| `Content-Type` | `application/json` | ประเภทข้อมูลที่ส่ง |
| `api-key` | `{api_key จาก config}` | API Key สำหรับ Authentication |

### 5.3 Request Body

**Source:** `controllers/missionController.go:387-394`

```json
{
    "log_id": "674f1234567890abcdef1234",
    "user_id": "U1234567890abcdef1234567890abcdef",
    "mission_detail": "Tier 1 Complete ตั้งแต่วันที่ 01/12/2025 10:00 - 08/12/2025 23:59 รวมยอดเดิมพันทั้งสิ้น 15000.00",
    "reward": 100.00,
    "callback_url": "https://your-server.com/api/missions/reward-callback",
    "line_at": "@lineofficial"
}
```

### 5.4 Request Body Fields Description

| Field | Type | Description |
|-------|------|-------------|
| `log_id` | string | MongoDB ObjectID ของ Log Entry ที่สร้างไว้ |
| `user_id` | string | LINE User ID ของ User |
| `mission_detail` | string | รายละเอียด Tier ที่ผ่าน (ช่วงวันที่ + ยอดเดิมพันรวม) |
| `reward` | float | จำนวนรางวัล (หน่วยเป็นเงิน) |
| `callback_url` | string | URL ที่ External API จะส่ง callback กลับมา |
| `line_at` | string | LINE Official Account ID |

### 5.5 Expected Response จาก External API

```json
// Success (200 OK หรือ 201 Created)
{
    "success": true,
    "message": "Reward claim received"
}
```

### 5.6 การสร้าง Log Entry ก่อนส่ง

**Source:** `controllers/missionController.go:368-384`

```go
logEntry := models.Log{
    UserID:        userID,
    MissionID:     missionID,
    MissionDetail: missionDetail,
    Reward:        reward,
    CreatedAt:     time.Now(),
    Status:        "pending",  // สถานะเริ่มต้น
}

result, err := c.logCollection.InsertOne(context.Background(), logEntry)
logID := result.InsertedID.(primitive.ObjectID)
```

**Log Entry Schema:**

```go
type Log struct {
    ID            primitive.ObjectID `bson:"_id,omitempty"`
    UserID        string             `bson:"user_id"`
    MissionID     string             `bson:"mission_id"`
    MissionDetail string             `bson:"mission_detail"`
    Reward        float64            `bson:"reward"`
    CreatedAt     time.Time          `bson:"created_at"`
    CallbackTime  time.Time          `bson:"callback_time,omitempty"`
    Status        string             `bson:"status"` // "pending", "approve", "reject"
}
```

---

## 6. ขั้นตอนที่ 4: External API Callback กลับมา

### 6.1 Callback API Endpoint

```
POST /api/missions/reward-callback
```

### 6.2 Request Body จาก External API

```json
{
    "log_id": "674f1234567890abcdef1234",
    "status": "approve"  // หรือ "reject"
}
```

### 6.3 Callback Flow

**Source:** `controllers/rewardCallbackController.go:35-127`

```go
func (c *RewardCallbackController) HandleRewardCallback(ctx *fiber.Ctx) error {
    // 1. Parse Request Body
    var callback struct {
        LogID  string `json:"log_id"`
        Status string `json:"status"` // "approve" or "reject"
    }

    // 2. ค้นหา Log Entry
    var logEntry models.Log
    logCollection.FindOne(bson.M{"_id": logID}).Decode(&logEntry)

    // 3. อัปเดต Log Status
    logCollection.UpdateOne(
        bson.M{"_id": logID, "status": "pending"},
        bson.M{"$set": bson.M{
            "callback_time": time.Now(),
            "status": callback.Status,
        }},
    )

    // 4. ดำเนินการตามสถานะ
    if callback.Status == "approve" {
        c.processSuccessfulReward(ctx.Context(), &mission)
    } else if callback.Status == "reject" {
        c.processFailedReward(ctx.Context(), &mission)
    }
}
```

### 6.4 กรณี Approve (อนุมัติ)

**Source:** `controllers/rewardCallbackController.go:129-187`

```go
func (c *RewardCallbackController) processSuccessfulReward(ctx context.Context, mission *models.Mission) error {
    currentTier.Status = "completed"

    if mission.CurrentTier < 3 {
        // Tier 1 หรือ 2: ไปยัง Tier ถัดไป
        mission.CurrentTier++
        newTier := createNewTier(newTierConfig)
        mission.Tiers = append(mission.Tiers, newTier)

        // สร้าง Events สำหรับ Tier ใหม่
        c.createNewEvents(...)
    } else {
        // Tier 3: สร้าง Level ใหม่ (ไม่จำกัด)
        currentTier.CurrentLevel++
        newLevel := createNewLevel(...)
        currentTier.Levels = append(currentTier.Levels, newLevel)
    }

    mission.Status = "processing"
    mission.ConsecutiveFails = 0
}
```

### 6.5 กรณี Reject (ปฏิเสธ)

**Source:** `controllers/rewardCallbackController.go:189-201`

```go
func (c *RewardCallbackController) processFailedReward(ctx context.Context, mission *models.Mission) error {
    currentTier.Status = "awaiting_reward"  // กลับไปสถานะรอรับรางวัล
    mission.Status = "processing"
}
```

### 6.6 Callback Response

```json
{
    "log_id": "674f1234567890abcdef1234",
    "callback_time": "2025-12-11T15:30:00Z",
    "status": "approve",
    "message": "Reward approved successfully"
}
```

---

## 7. External API Endpoints ทั้งหมด

### 7.1 Phone Number Sync (ตรวจสอบเบอร์โทร)

**Source:** `controllers/clientController.go:285-349`

```
POST {api_endpoint}/players/v1/line/sync
```

**Headers:**
```
Content-Type: application/json
API-KEY: {api_key}
```

**Request:**
```json
{
    "phone_number": "0812345678",
    "line_id": "U1234567890abcdef...",
    "line_at": "@lineofficial"
}
```

**Response (Success):**
```json
{
    "username": "0812345678"
}
```

---

### 7.2 Get Bets (ดึงยอดเดิมพัน)

**Source:** `utils/getCurrentBet.go:14-75`

```
GET {api_endpoint}/players/v1/line/bets?line_id={LINE_USER_ID}&line_at={LINE_AT}&start_date={UNIX_TIMESTAMP}&end_date={UNIX_TIMESTAMP}
```

**Headers:**
```
API-KEY: {api_key}
```

**Example URL:**
```
GET https://example.com/api/players/v1/line/bets?line_id=U1234...&line_at=@lineofficial&start_date=1733097600&end_date=1733702399
```

**Response:**
```json
{
    "bet": 15000.00
}
```

---

### 7.3 Claim Reward (ส่งคำขอรับรางวัล)

**Source:** `controllers/missionController.go:367-435`

```
POST {api_endpoint}/players/v1/line/rewards/claim
```

**Headers:**
```
Content-Type: application/json
api-key: {api_key}
```

**Request:**
```json
{
    "log_id": "674f1234567890abcdef1234",
    "user_id": "U1234567890abcdef...",
    "mission_detail": "Tier 1 Complete...",
    "reward": 100.00,
    "callback_url": "https://your-server.com/api/missions/reward-callback",
    "line_at": "@lineofficial"
}
```

---

## 8. Authentication

### 8.1 API Key Authentication

ทุก Request ที่ส่งไปยัง External API ต้องมี API Key ใน Header:

| Endpoint | Header Name | Value |
|----------|-------------|-------|
| `/players/v1/line/sync` | `API-KEY` | `{config.ApiKey}` |
| `/players/v1/line/bets` | `API-KEY` | `{config.ApiKey}` |
| `/players/v1/line/rewards/claim` | `api-key` | `{config.ApiKey}` |

> **หมายเหตุ:** Header name ใช้ตัวพิมพ์ต่างกัน (`API-KEY` vs `api-key`) ควรตรวจสอบกับ External API

### 8.2 Config Storage

API Key และ Endpoint ถูกเก็บใน MongoDB Collection `tbl_config`:

```json
{
    "api_endpoint": "https://example.com/api",
    "api_key": "secret-api-key-here",
    "line_at": "@lineofficial"
}
```

---

## 9. Error Handling

### 9.1 External API Errors

```go
// จาก controllers/missionController.go:430-432
if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
    return fmt.Errorf("external API returned non-OK status: %d", resp.StatusCode)
}
```

### 9.2 Get Bet API Fallback

```go
// จาก controllers/missionController.go:168-172
currentBet, err := utils.GetCurrentBet(config, mission.UserID, startDate, endDate)
if err != nil {
    log.Printf("Failed to get current bet: %v", err)
    currentBet = 0 // Fallback to 0 if API call fails
}
```

### 9.3 Notification Failures

การส่งแจ้งเตือน (LINE, Telegram) หากล้มเหลวจะไม่หยุด Flow หลัก:

```go
// จาก controllers/missionController.go:292-296
err = c.telegramController.SendRewardClaimedMessage(...)
if err != nil {
    log.Printf("Failed to send Telegram message: %v", err)
    // Continue with the process even if sending the message fails
}
```

---

## Summary Table

| Step | Action | API Endpoint | Direction |
|------|--------|--------------|-----------|
| 1 | User ผ่านครบ 3 Levels | - | Internal |
| 2 | User กด Claim Reward | `POST /api/missions/:id/claim-reward` | Client → Server |
| 3 | สร้าง Log Entry | - | Internal (MongoDB) |
| 4 | ส่งคำขอไป External | `POST {api}/players/v1/line/rewards/claim` | Server → External |
| 5 | External ประมวลผล | - | External System |
| 6 | External Callback | `POST /api/missions/reward-callback` | External → Server |
| 7 | อัปเดต Mission | - | Internal (MongoDB) |

---

*Document generated: 2025-12-11*
*เอกสารสร้างโดย: Claude Code*
