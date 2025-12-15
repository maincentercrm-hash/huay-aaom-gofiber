# 📊 Data Models Documentation

## Overview
ระบบประกอบด้วย 8 models หลักสำหรับจัดการข้อมูลของระบบ Retirement Lottery

---

## 🧑‍💼 User Model
**File:** `models/user.go`
**Collection:** `tbl_users`

```go
type User struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Email      string             `bson:"email" json:"email"`
    Password   string             `bson:"password" json:"-"`
    Role       string             `bson:"role" json:"role"`
    Status     string             `bson:"status" json:"status"`
    CreateDate time.Time          `bson:"createDate" json:"createDate"`
}
```

**Purpose:** จัดการข้อมูลผู้ใช้แอดมิน
**Key Features:**
- Password ไม่ส่งกลับใน JSON response (json:"-")
- Support role-based access control
- Track user status และวันที่สร้าง

**API Endpoints:**
- POST `/api/admin/login` - Admin login
- POST `/api/admin/register` - Admin registration

---

## 👥 Client Model
**File:** `models/client.go`
**Collection:** `tbl_client`

```go
type Client struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID        string             `bson:"user_id" json:"userId"`
    DisplayName   string             `bson:"display_name" json:"displayName"`
    PictureURL    string             `bson:"picture_url" json:"pictureUrl"`
    StatusMessage string             `bson:"status_message" json:"statusMessage"`
    PhoneNumber   string             `bson:"phone_number" json:"phoneNumber,omitempty"`
    CreatedAt     primitive.DateTime `bson:"created_at" json:"createdAt"`
    UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updatedAt"`
}
```

**Purpose:** จัดการข้อมูล LINE users ที่เข้าใช้ระบบ
**Key Features:**
- UserID เป็น LINE User ID
- รองรับข้อมูลโปรไฟล์จาก LINE (display name, picture, status)
- Track phone number สำหรับการตรวจสอบ
- Auto timestamp (CreatedAt, UpdatedAt)

**API Endpoints:**
- GET `/api/clients/` - ดูรายชื่อ client ทั้งหมด (มี pagination และ search)
- POST `/api/clients/` - สร้าง/อัปเดต client
- GET `/api/clients/:id` - ดูข้อมูล client (รองรับทั้ง ObjectID และ user_id)
- DELETE `/api/clients/:id` - ลบ client (รองรับทั้ง ObjectID และ user_id)
- GET `/api/clients/:userId/check-phone` - ตรวจสอบเบอร์โทร
- PUT `/api/clients/:userId/update-phone` - อัปเดตเบอร์โทร

---

## 💰 UserBet Model
**File:** `models/userBet.go`
**Collection:** `tbl_user_bet`

```go
type UserBet struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID     string             `bson:"user_id" json:"user_id"`
    CurrentBet float64            `bson:"current_bet" json:"current_bet"`
    UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
```

**Purpose:** ติดตามยอดเงินเดิมพันปัจจุบันของแต่ละ user
**Key Features:**
- เก็บยอดเดิมพันล่าสุด
- Auto update timestamp
- Link กับ user ผ่าน UserID (LINE User ID)

**API Endpoints:**
- Generic routes ผ่าน `/api/user_bet/`

---

## 🎯 Mission Model
**File:** `models/mission.go`
**Collection:** `tbl_mission`

**Main Model:**
```go
type Mission struct {
    ID               primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID           string             `bson:"user_id" json:"user_id"`
    PhoneNumber      string             `bson:"phone_number" json:"phone_number"`
    Status           string             `bson:"status" json:"status"`
    CurrentTier      int                `bson:"current_tier" json:"current_tier"`
    Tiers            []Tier             `bson:"tiers" json:"tiers"`
    ConsecutiveFails int                `bson:"consecutive_fails" json:"consecutive_fails"`
    CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}
```

**Sub-Models:**
```go
type Tier struct {
    Name         string    `bson:"name" json:"name"`
    Reward       int       `bson:"reward" json:"reward"`
    Target       int       `bson:"target" json:"target"`
    Status       string    `bson:"status" json:"status"`
    CurrentLevel int       `bson:"current_level" json:"current_level"`
    MaxLevel     int       `bson:"max_level" json:"max_level"`
    Levels       []Level   `bson:"levels" json:"levels"`
    ExpireReward time.Time `bson:"expire_reward" json:"expire_reward"`
}

type Level struct {
    Name         string    `bson:"name" json:"name"`
    StartDate    time.Time `bson:"start_date" json:"start_date"`
    ExpireDate   time.Time `bson:"expire_date" json:"expire_date"`
    FollowUpDate time.Time `bson:"follow_up_date" json:"follow_up_date"`
    Status       string    `bson:"status" json:"status"`
    CurrentBet   float64   `bson:"current_bet" json:"current_bet"`
}
```

**Purpose:** จัดการระบบภารกิจ (missions) แบบหลายระดับ
**Key Features:**
- ระบบ Tier และ Level แบบซ้อนกัน
- ติดตาม status ของแต่ละระดับ
- Track consecutive fails
- วันหมดอายุและ follow-up dates
- เก็บยอดเดิมพันแต่ละ level

**Status Values:**
- Mission: "processing", "completed", "failed", "pending"
- Tier/Level: ตามการกำหนดของระบบ

**API Endpoints:**
- `/api/missions/` - Mission management endpoints

---

## ⚙️ Config Model
**File:** `models/config.go`
**Collection:** `tbl_config`

**Main Model:**
```go
type Config struct {
    ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    LiffID             string             `bson:"liff_id" json:"liff_id"`
    ChannelAccessToken string             `bson:"channel_access_token" json:"channel_access_token"`
    ChannelSecret      string             `bson:"channel_secret" json:"channel_secret"`
    Tiers              []TierDetail       `bson:"tiers" json:"tiers"`
    TelegramBotToken   string             `bson:"telegram_bot_token" json:"telegram_bot_token"`
    TelegramChatID     string             `bson:"telegram_chat_id" json:"telegram_chat_id"`
    FirebaseConfig     FirebaseConfig     `bson:"firebase_config" json:"firebase_config"`
    FlexMessages       FlexMessages       `bson:"flex_messages" json:"flexMessages"`
    SiteTemplate       SiteTemplateConfig `bson:"site_template" json:"siteTemplate"`
    ApiEndpoint        string             `bson:"api_endpoint" json:"api_endpoint"`
    ApiKey             string             `bson:"api_key" json:"api_key"`
    LineAt             string             `bson:"line_at" json:"line_at"`
    LineSyncURL        string             `bson:"line_sync_url" json:"line_sync_url"`
}
```

**Sub-Models รวม 15+ structures:**
- `TierDetail` - กำหนดค่า tier แต่ละระดับ
- `FirebaseConfig` - การตั้งค่า Firebase
- `FlexMessages` - เทมเพลตข้อความ LINE Flex
- `SiteTemplateConfig` - การตั้งค่าหน้าเว็บ
- `TierConfig` - สีและธีมของแต่ละ tier
- และอีกมากมาย...

**Purpose:** จัดการการตั้งค่าทั้งระบบ
**Key Features:**
- LINE API configuration
- Telegram integration
- Firebase settings
- Flex message templates
- Site appearance customization
- Mission tier configurations
- External API endpoints

**API Endpoints:**
- `/api/config/` - Configuration management

---

## 📨 MessageLog Model
**File:** `models/message.go`
**Collection:** `tbl_logs_message`

```go
type MessageLog struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID      string             `bson:"user_id" json:"user_id"`
    Status      string             `bson:"status" json:"status"`
    Tier        string             `bson:"tier" json:"tier"`
    Level       string             `bson:"level" json:"level"`
    MissionID   primitive.ObjectID `bson:"mission_id" json:"mission_id"`
    SentAt      time.Time          `bson:"sent_at" json:"sent_at"`
    ReadAt      time.Time          `bson:"read_at,omitempty" json:"read_at,omitempty"`
    FlexContent FlexContent        `bson:"flex_content" json:"flex_content"`
}

type FlexContent struct {
    Title          string `bson:"title" json:"title"`
    Description    string `bson:"description" json:"description"`
    SubDescription string `bson:"sub_description,omitempty" json:"sub_description,omitempty"`
}
```

**Purpose:** บันทึกประวัติการส่งข้อความ
**Key Features:**
- ติดตาม status ข้อความ ("sent", "read", "unread")
- Link กับ mission และ user
- เก็บเนื้อหา Flex message
- Track การอ่านข้อความ (ReadAt)

**API Endpoints:**
- Generic routes ผ่าน `/api/logs_message/`

---

## 📊 Log Model
**File:** `models/log.go`
**Collection:** `tbl_logs`

```go
type Log struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID        string             `bson:"user_id" json:"user_id"`
    MissionID     string             `bson:"mission_id" json:"mission_id"`
    MissionDetail string             `bson:"mission_detail" json:"mission_detail"`
    Reward        float64            `bson:"reward" json:"reward"`
    CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
    CallbackTime  time.Time          `bson:"callback_time,omitempty" json:"callback_time,omitempty"`
    Status        string             `bson:"status" json:"status"`
}
```

**Purpose:** บันทึกประวัติการทำภารกิจและรางวัล
**Key Features:**
- เก็บรายละเอียดภารกิจ
- ติดตามยอดรางวัล
- Status tracking ("pending", "approve", "reject")
- Callback time สำหรับการตอบกลับ

**API Endpoints:**
- ไม่มี specific endpoints (ใช้สำหรับ logging internal)

---

## ⏰ ExpirationEvent Model
**File:** `models/expirationEvent.go`
**Collection:** `tbl_expiration_events`

```go
type ExpirationEvent struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    MissionID  primitive.ObjectID `bson:"mission_id" json:"mission_id"`
    TierIndex  int                `bson:"tier_index" json:"tier_index"`
    LevelIndex int                `bson:"level_index" json:"level_index"`
    ExpireTime time.Time          `bson:"expire_time" json:"expire_time"`
    Status     string             `bson:"status" json:"status"`
    Type       string             `bson:"type" json:"type"`
}
```

**Purpose:** จัดการกิจกรรมที่มีวันหมดอายุ
**Key Features:**
- ติดตาม mission/tier/level ที่จะหมดอายุ
- Type หลายแบบ: "level_expiration", "follow_up", "reward_expiration"
- Status: "pending", "processed"
- ใช้สำหรับ cron jobs และการแจ้งเตือน

**API Endpoints:**
- ไม่มี public endpoints (ใช้ internal)

---

## 🗂️ Collections Summary

| Model | Collection | Purpose |
|-------|------------|---------|
| User | `tbl_users` | Admin users |
| Client | `tbl_client` | LINE users |
| UserBet | `tbl_user_bet` | Current bet amounts |
| Mission | `tbl_mission` | Mission progress |
| Config | `tbl_config` | System configuration |
| MessageLog | `tbl_logs_message` | Message history |
| Log | `tbl_logs` | Mission logs |
| ExpirationEvent | `tbl_expiration_events` | Expiration tracking |

---

## 🔄 Relationships

```
User (Admin) ←→ System Management
    ↓
Client (LINE Users) ←→ UserBet ←→ Mission
    ↓                      ↓         ↓
MessageLog ←→ Config ←→ ExpirationEvent
    ↓
Log (Mission History)
```

**Key Relationships:**
- **Client ↔ UserBet:** 1:1 (UserID)
- **Client ↔ Mission:** 1:N (UserID)
- **Mission ↔ MessageLog:** 1:N (MissionID)
- **Mission ↔ ExpirationEvent:** 1:N (MissionID)
- **Config:** Global singleton (1 record)

---

## 📝 Notes

### Security Considerations:
- User.Password จะไม่ส่งกลับใน JSON
- Config มี sensitive data (tokens, secrets)
- Client endpoints รองรับทั้ง ObjectID และ user_id

### Data Types:
- ใช้ `primitive.ObjectID` สำหรับ MongoDB _id
- ใช้ `primitive.DateTime` สำหรับ MongoDB datetime
- ใช้ `time.Time` สำหรับ Go native datetime

### API Coverage:
- **Full CRUD:** Client, Config
- **Generic CRUD:** User, UserBet, MessageLog
- **Custom Logic:** Mission
- **Internal Only:** Log, ExpirationEvent