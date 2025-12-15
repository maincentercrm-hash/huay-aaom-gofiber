# Testing Checklist - Retirement Lottery System
# รายการทดสอบระบบ - ระบบสลากเกษียณ

> **Version:** 1.0
> **Last Updated:** 2025-12-08
> **Framework:** Go Fiber v2 + MongoDB
> **External Services:** LINE API, Telegram API, Firebase Storage, External Players API

---

## สารบัญ (Table of Contents)

1. [Admin Authentication - ระบบยืนยันตัวตนผู้ดูแล](#1-admin-authentication---ระบบยืนยันตัวตนผู้ดูแล)
2. [Client Management - ระบบจัดการลูกค้า](#2-client-management---ระบบจัดการลูกค้า)
3. [Mission Management - ระบบจัดการภารกิจ](#3-mission-management---ระบบจัดการภารกิจ)
4. [Configuration Management - ระบบจัดการการตั้งค่า](#4-configuration-management---ระบบจัดการการตั้งค่า)
5. [Dashboard - แดชบอร์ด](#5-dashboard---แดชบอร์ด)
6. [User Bet - ระบบยอดเดิมพันผู้ใช้](#6-user-bet---ระบบยอดเดิมพันผู้ใช้)
7. [LINE Messaging - ระบบส่งข้อความ LINE](#7-line-messaging---ระบบส่งข้อความ-line)
8. [Telegram Notification - ระบบแจ้งเตือน Telegram](#8-telegram-notification---ระบบแจ้งเตือน-telegram)
9. [Background Jobs - งานเบื้องหลัง](#9-background-jobs---งานเบื้องหลัง-event-processing)
10. [External API Integration - การเชื่อมต่อ API ภายนอก](#10-external-api-integration---การเชื่อมต่อ-api-ภายนอก)
11. [File Upload - ระบบอัปโหลดไฟล์](#11-file-upload---ระบบอัปโหลดไฟล์-firebase)
12. [Generic CRUD Operations - การดำเนินการ CRUD ทั่วไป](#12-generic-crud-operations---การดำเนินการ-crud-ทั่วไป)
13. [Security Testing - การทดสอบความปลอดภัย](#13-security-testing---การทดสอบความปลอดภัย)
14. [Performance Testing - การทดสอบประสิทธิภาพ](#14-performance-testing---การทดสอบประสิทธิภาพ)
15. [Database & Data Integrity - ฐานข้อมูลและความถูกต้องของข้อมูล](#15-database--data-integrity---ฐานข้อมูลและความถูกต้องของข้อมูล)
16. [Error Handling - การจัดการข้อผิดพลาด](#16-error-handling---การจัดการข้อผิดพลาด)
17. [Edge Cases & Boundary Testing - กรณีขอบเขตและการทดสอบขีดจำกัด](#17-edge-cases--boundary-testing---กรณีขอบเขตและการทดสอบขีดจำกัด)

---

## 1. Admin Authentication - ระบบยืนยันตัวตนผู้ดูแล

### 1.1 Login - เข้าสู่ระบบ (`POST /api/admin/login`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 1.1.1 | Login with valid email and password | เข้าสู่ระบบด้วยอีเมลและรหัสผ่านที่ถูกต้อง | Return JWT token + user info (200) | [ ] |
| 1.1.2 | Login with invalid email | เข้าสู่ระบบด้วยอีเมลที่ไม่มีในระบบ | Return error "User not found" (404) | [ ] |
| 1.1.3 | Login with wrong password | เข้าสู่ระบบด้วยรหัสผ่านผิด | Return error "Invalid password" (401) | [ ] |
| 1.1.4 | Login with empty email | เข้าสู่ระบบโดยไม่กรอกอีเมล | Return error (400) | [ ] |
| 1.1.5 | Login with empty password | เข้าสู่ระบบโดยไม่กรอกรหัสผ่าน | Return error (400) | [ ] |
| 1.1.6 | Login with SQL injection in email | ทดสอบ SQL injection ในช่องอีเมล | Return error, no data leak | [ ] |
| 1.1.7 | Verify JWT token contains correct claims | ตรวจสอบว่า JWT token มี claims ถูกต้อง (user_id, role, exp) | Claims match user data | [ ] |
| 1.1.8 | Verify JWT expires after 12 hours | ตรวจสอบว่า JWT หมดอายุหลัง 12 ชั่วโมง | Token invalid after 12 hours | [ ] |
| 1.1.9 | Login with case-sensitive email | ทดสอบอีเมลตัวพิมพ์ใหญ่-เล็ก | Verify email case handling | [ ] |

### 1.2 Register - ลงทะเบียน (`POST /api/admin/register`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 1.2.1 | Register with valid data | ลงทะเบียนด้วยข้อมูลถูกต้อง (email, password, role) | Return JWT token + user info (201) | [ ] |
| 1.2.2 | Register with existing email | ลงทะเบียนด้วยอีเมลที่มีอยู่แล้ว | Return error "Email already exists" | [ ] |
| 1.2.3 | Register with empty email | ลงทะเบียนโดยไม่กรอกอีเมล | Return error (400) | [ ] |
| 1.2.4 | Register with empty password | ลงทะเบียนโดยไม่กรอกรหัสผ่าน | Return error (400) | [ ] |
| 1.2.5 | Register with invalid email format | ลงทะเบียนด้วยรูปแบบอีเมลไม่ถูกต้อง | Return error (400) | [ ] |
| 1.2.6 | Verify password is hashed with bcrypt | ตรวจสอบว่ารหัสผ่านถูกเข้ารหัสด้วย bcrypt | Password not stored in plaintext | [ ] |
| 1.2.7 | Register with valid roles | ลงทะเบียนด้วย role ที่ถูกต้อง (admin, user) | Accept valid roles only | [ ] |
| 1.2.8 | Register with invalid role | ลงทะเบียนด้วย role ที่ไม่ถูกต้อง | Return error or default role | [ ] |

---

## 2. Client Management - ระบบจัดการลูกค้า

### 2.1 Get All Clients - ดึงข้อมูลลูกค้าทั้งหมด (`GET /api/clients`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 2.1.1 | Get all clients without pagination | ดึงลูกค้าทั้งหมดโดยไม่ระบุ pagination | Return clients with default pagination | [ ] |
| 2.1.2 | Get clients with page=1, limit=10 | ดึงลูกค้าหน้าแรก 10 รายการ | Return first 10 clients | [ ] |
| 2.1.3 | Get clients with page=2, limit=10 | ดึงลูกค้าหน้าที่ 2 (รายการที่ 11-20) | Return clients 11-20 | [ ] |
| 2.1.4 | Get clients with search query | ค้นหาลูกค้าด้วยคำค้น | Return matching clients | [ ] |
| 2.1.5 | Get clients with empty search | ค้นหาด้วยคำค้นว่าง | Return all clients | [ ] |
| 2.1.6 | Get clients when no data exists | ดึงข้อมูลเมื่อไม่มีลูกค้าในระบบ | Return empty array with pagination | [ ] |
| 2.1.7 | Verify pagination metadata | ตรวจสอบข้อมูล pagination (total_count, has_next, has_prev) | Metadata is accurate | [ ] |
| 2.1.8 | Get clients with limit=0 | ดึงข้อมูลด้วย limit=0 | Handle gracefully (default limit) | [ ] |
| 2.1.9 | Get clients with negative page | ดึงข้อมูลด้วยหน้าติดลบ | Handle gracefully | [ ] |
| 2.1.10 | Get clients with very large limit | ดึงข้อมูลด้วย limit มากเกินไป (>100) | Limit capped at max | [ ] |

### 2.2 Upsert Client - สร้างหรืออัปเดตลูกค้า (`POST /api/clients`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 2.2.1 | Create new client with valid data | สร้างลูกค้าใหม่ด้วยข้อมูลถูกต้อง | Return created client (201) | [ ] |
| 2.2.2 | Update existing client | อัปเดตลูกค้าที่มีอยู่แล้ว | Return updated client (200) | [ ] |
| 2.2.3 | Create client with LINE user data | สร้างลูกค้าด้วยข้อมูลจาก LINE | Save all LINE fields | [ ] |
| 2.2.4 | Create client with empty user_id | สร้างลูกค้าโดยไม่มี user_id | Return error (400) | [ ] |
| 2.2.5 | Verify CreatedAt is set on create | ตรวจสอบว่า CreatedAt ถูกตั้งค่าเมื่อสร้าง | Timestamp is set | [ ] |
| 2.2.6 | Verify UpdatedAt is updated on update | ตรวจสอบว่า UpdatedAt ถูกอัปเดตเมื่อแก้ไข | Timestamp is updated | [ ] |
| 2.2.7 | Create client with special characters | สร้างลูกค้าด้วยชื่อที่มีอักขระพิเศษ/ภาษาไทย | Handle unicode properly | [ ] |

### 2.3 Get Client by ID - ดึงข้อมูลลูกค้าตาม ID (`GET /api/clients/:userId`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 2.3.1 | Get client with valid ObjectID | ดึงลูกค้าด้วย ObjectID ที่ถูกต้อง | Return client data | [ ] |
| 2.3.2 | Get client with valid LINE user_id | ดึงลูกค้าด้วย LINE user_id | Return client data | [ ] |
| 2.3.3 | Get client with invalid ID | ดึงลูกค้าด้วย ID ที่รูปแบบไม่ถูกต้อง | Return error (400) | [ ] |
| 2.3.4 | Get client with non-existent ID | ดึงลูกค้าด้วย ID ที่ไม่มีในระบบ | Return error (404) | [ ] |
| 2.3.5 | Get client with empty ID | ดึงลูกค้าโดยไม่ระบุ ID | Return error (400) | [ ] |

### 2.4 Delete Client - ลบลูกค้า (`DELETE /api/clients/:userId`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 2.4.1 | Delete existing client | ลบลูกค้าที่มีอยู่ในระบบ | Return success message | [ ] |
| 2.4.2 | Delete non-existent client | ลบลูกค้าที่ไม่มีในระบบ | Return error (404) | [ ] |
| 2.4.3 | Delete client with invalid ID format | ลบลูกค้าด้วย ID รูปแบบไม่ถูกต้อง | Return error (400) | [ ] |
| 2.4.4 | Verify related missions are handled | ตรวจสอบว่าภารกิจที่เกี่ยวข้องถูกจัดการ | Check data consistency | [ ] |

### 2.5 Check Phone Number - ตรวจสอบเบอร์โทร (`GET /api/clients/:userId/check-phone`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 2.5.1 | Check client with phone number | ตรวจสอบลูกค้าที่มีเบอร์โทร | Return {hasPhoneNumber: true} | [ ] |
| 2.5.2 | Check client without phone number | ตรวจสอบลูกค้าที่ไม่มีเบอร์โทร | Return {hasPhoneNumber: false} | [ ] |
| 2.5.3 | Check non-existent client | ตรวจสอบลูกค้าที่ไม่มีในระบบ | Return error (404) | [ ] |
| 2.5.4 | Check with empty phone number field | ตรวจสอบเมื่อช่องเบอร์โทรว่าง | Return {hasPhoneNumber: false} | [ ] |

### 2.6 Update Phone Number - อัปเดตเบอร์โทร (`PUT /api/clients/:userId/update-phone`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 2.6.1 | Update with valid phone number | อัปเดตด้วยเบอร์โทรที่ถูกต้อง | Return success, phone verified via external API | [ ] |
| 2.6.2 | Update with invalid phone format | อัปเดตด้วยรูปแบบเบอร์โทรไม่ถูกต้อง | Return error (400) | [ ] |
| 2.6.3 | Update with phone not matching external system | อัปเดตด้วยเบอร์โทรที่ไม่ตรงกับระบบภายนอก | Return error from API | [ ] |
| 2.6.4 | Update non-existent client | อัปเดตลูกค้าที่ไม่มีในระบบ | Return error (404) | [ ] |
| 2.6.5 | Update with empty phone number | อัปเดตด้วยเบอร์โทรว่าง | Return error (400) | [ ] |
| 2.6.6 | Verify external API sync is called | ตรวจสอบว่า API ภายนอกถูกเรียก | API is called with correct params | [ ] |

---

## 3. Mission Management - ระบบจัดการภารกิจ

### 3.1 Create Mission - สร้างภารกิจ (`POST /api/missions`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 3.1.1 | Create mission with valid data | สร้างภารกิจด้วย user_id และ phone_number ที่ถูกต้อง | Return mission with Tier 1, Level 1 | [ ] |
| 3.1.2 | Create mission with empty user_id | สร้างภารกิจโดยไม่มี user_id | Return error (400) | [ ] |
| 3.1.3 | Create mission with empty phone_number | สร้างภารกิจโดยไม่มี phone_number | Return error (400) | [ ] |
| 3.1.4 | Create mission when user already has active mission | สร้างภารกิจเมื่อผู้ใช้มีภารกิจที่ยังดำเนินอยู่ | Return error (409) | [ ] |
| 3.1.5 | Verify initial status is "processing" | ตรวจสอบสถานะเริ่มต้นเป็น "processing" | Status = processing | [ ] |
| 3.1.6 | Verify Tier 1 is created with correct config | ตรวจสอบว่า Tier 1 ถูกสร้างตาม config | Tier matches config | [ ] |
| 3.1.7 | Verify Level 1 is created with correct dates | ตรวจสอบว่า Level 1 มีวันที่ถูกต้อง | Dates calculated correctly | [ ] |
| 3.1.8 | Verify level_expiration event is created | ตรวจสอบว่า event หมดอายุ level ถูกสร้าง | Event exists in tbl_events | [ ] |
| 3.1.9 | Verify follow_up event is created | ตรวจสอบว่า event ติดตามผลถูกสร้าง | Event exists in tbl_events | [ ] |
| 3.1.10 | Verify consecutive_fails initialized to 0 | ตรวจสอบว่าจำนวนล้มเหลวติดต่อกันเริ่มที่ 0 | consecutive_fails = 0 | [ ] |

### 3.2 Update Mission Status - อัปเดตสถานะภารกิจ (`PUT /api/missions/:id/status`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 3.2.1 | Update when bet >= target (level success) | อัปเดตเมื่อยอดเดิมพัน >= เป้าหมาย (ผ่าน) | Level status = completed | [ ] |
| 3.2.2 | Update when bet < target (level fail) | อัปเดตเมื่อยอดเดิมพัน < เป้าหมาย (ไม่ผ่าน) | consecutive_fails incremented | [ ] |
| 3.2.3 | Update with max consecutive fails reached | อัปเดตเมื่อล้มเหลวติดต่อกันครบจำนวนสูงสุด | Mission status = failed | [ ] |
| 3.2.4 | Update completing last level of tier | อัปเดตเมื่อผ่าน level สุดท้ายของ tier | Tier status = completed | [ ] |
| 3.2.5 | Update with invalid mission ID | อัปเดตด้วย mission ID ไม่ถูกต้อง | Return error (400) | [ ] |
| 3.2.6 | Update non-existent mission | อัปเดตภารกิจที่ไม่มีในระบบ | Return error (404) | [ ] |
| 3.2.7 | Update with invalid tierIndex | อัปเดตด้วย tierIndex ไม่ถูกต้อง | Return error (400) | [ ] |
| 3.2.8 | Verify new level created after success | ตรวจสอบว่า level ใหม่ถูกสร้างหลังผ่าน | New level exists | [ ] |
| 3.2.9 | Verify new tier created after tier completion | ตรวจสอบว่า tier ใหม่ถูกสร้างหลังจบ tier (Tier 1/2) | New tier exists | [ ] |
| 3.2.10 | Verify events created for new level | ตรวจสอบว่า events ถูกสร้างสำหรับ level ใหม่ | Events exist | [ ] |
| 3.2.11 | Verify LINE message sent on success | ตรวจสอบว่าส่งข้อความ LINE เมื่อผ่าน | Message logged | [ ] |
| 3.2.12 | Verify LINE message sent on failure | ตรวจสอบว่าส่งข้อความ LINE เมื่อไม่ผ่าน | Message logged | [ ] |

### 3.3 Get Processing Mission - ดึงภารกิจที่กำลังดำเนินอยู่ (`GET /api/missions/processing`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 3.3.1 | Get mission for user with active mission | ดึงภารกิจสำหรับผู้ใช้ที่มีภารกิจดำเนินอยู่ | Return mission data | [ ] |
| 3.3.2 | Get mission for user without active mission | ดึงภารกิจสำหรับผู้ใช้ที่ไม่มีภารกิจ | Return error (404) | [ ] |
| 3.3.3 | Get mission with empty user_id query | ดึงภารกิจโดยไม่ระบุ user_id | Return error (400) | [ ] |
| 3.3.4 | Get mission with invalid user_id format | ดึงภารกิจด้วย user_id รูปแบบไม่ถูกต้อง | Handle gracefully | [ ] |
| 3.3.5 | Verify only "processing" status returned | ตรวจสอบว่าคืนเฉพาะสถานะ "processing" | No failed/completed missions | [ ] |

### 3.4 Claim Reward - ขอรับรางวัล (`POST /api/missions/:id/claim-reward`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 3.4.1 | Claim reward when tier is "completed" | ขอรางวัลเมื่อ tier สถานะ "completed" | Return status pending | [ ] |
| 3.4.2 | Claim reward when tier is "awaiting_reward" | ขอรางวัลเมื่อ tier สถานะ "awaiting_reward" | Return status pending | [ ] |
| 3.4.3 | Claim reward when tier not completed | ขอรางวัลเมื่อ tier ยังไม่เสร็จ | Return error (400) | [ ] |
| 3.4.4 | Claim with invalid mission ID | ขอรางวัลด้วย mission ID ไม่ถูกต้อง | Return error (400) | [ ] |
| 3.4.5 | Claim non-existent mission | ขอรางวัลภารกิจที่ไม่มีในระบบ | Return error (404) | [ ] |
| 3.4.6 | Verify log entry created with status "pending" | ตรวจสอบว่า log ถูกสร้างด้วยสถานะ "pending" | Log exists in tbl_logs | [ ] |
| 3.4.7 | Verify external API called | ตรวจสอบว่า API ภายนอกถูกเรียก | API request sent | [ ] |
| 3.4.8 | Verify LINE notification sent | ตรวจสอบว่าส่งแจ้งเตือน LINE | Message logged | [ ] |
| 3.4.9 | Verify Telegram notification sent | ตรวจสอบว่าส่งแจ้งเตือน Telegram ให้แอดมิน | Admin notified | [ ] |

### 3.5 Check Existing Mission - ตรวจสอบภารกิจที่มีอยู่ (`GET /api/missions/check`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 3.5.1 | Check user with active mission | ตรวจสอบผู้ใช้ที่มีภารกิจดำเนินอยู่ | Return {hasMission: true} | [ ] |
| 3.5.2 | Check user without mission | ตรวจสอบผู้ใช้ที่ไม่มีภารกิจ | Return {hasMission: false} | [ ] |
| 3.5.3 | Check with empty user_id | ตรวจสอบโดยไม่ระบุ user_id | Return error (400) | [ ] |
| 3.5.4 | Check with completed mission | ตรวจสอบผู้ใช้ที่มีภารกิจเสร็จสิ้นแล้ว | Return {hasMission: false} | [ ] |
| 3.5.5 | Check with failed mission | ตรวจสอบผู้ใช้ที่มีภารกิจล้มเหลว | Return {hasMission: false} | [ ] |

### 3.6 Reward Callback - รับผลการอนุมัติรางวัล (`POST /api/missions/reward-callback`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 3.6.1 | Callback with status "approve" | รับ callback สถานะ "อนุมัติ" | Tier completed, move to next | [ ] |
| 3.6.2 | Callback with status "reject" | รับ callback สถานะ "ปฏิเสธ" | Tier stays awaiting_reward | [ ] |
| 3.6.3 | Callback with invalid log_id | รับ callback ด้วย log_id ไม่ถูกต้อง | Return error (400) | [ ] |
| 3.6.4 | Callback with non-existent log | รับ callback ด้วย log ที่ไม่มีในระบบ | Return error (404) | [ ] |
| 3.6.5 | Callback with invalid status | รับ callback ด้วยสถานะไม่ถูกต้อง | Return error (400) | [ ] |
| 3.6.6 | Verify log status updated | ตรวจสอบว่า log status ถูกอัปเดต | Log status matches | [ ] |
| 3.6.7 | Verify LINE notification sent on approve | ตรวจสอบว่าส่งแจ้งเตือน LINE เมื่ออนุมัติ | Message logged | [ ] |
| 3.6.8 | Verify new tier created after approve | ตรวจสอบว่า tier ใหม่ถูกสร้างหลังอนุมัติ (Tier 1/2) | New tier exists | [ ] |

---

## 4. Configuration Management - ระบบจัดการการตั้งค่า

### 4.1 Get Config - ดึงการตั้งค่า (`GET /api/config`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 4.1.1 | Get existing config | ดึงการตั้งค่าที่มีอยู่ | Return full config object | [ ] |
| 4.1.2 | Get when no config exists | ดึงเมื่อไม่มีการตั้งค่า | Return empty/default config | [ ] |
| 4.1.3 | Verify all fields returned | ตรวจสอบว่าคืนทุก field (LIFF, LINE, Telegram, Tiers, etc.) | All fields present | [ ] |
| 4.1.4 | Verify sensitive data handling | ตรวจสอบการจัดการข้อมูลลับ | No plaintext secrets in response | [ ] |

### 4.2 Save Config - บันทึกการตั้งค่า (`POST /api/config`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 4.2.1 | Save complete config | บันทึกการตั้งค่าทั้งหมด | Return updated config | [ ] |
| 4.2.2 | Save with missing required fields | บันทึกโดยขาด field จำเป็น | Handle gracefully | [ ] |
| 4.2.3 | Save LINE credentials | บันทึกข้อมูล LINE credentials | Credentials saved | [ ] |
| 4.2.4 | Save Telegram credentials | บันทึกข้อมูล Telegram credentials | Credentials saved | [ ] |
| 4.2.5 | Save tier configuration | บันทึกการตั้งค่า tier | Tiers saved correctly | [ ] |
| 4.2.6 | Verify existing config updated | ตรวจสอบว่าอัปเดต config ที่มีอยู่ (ไม่สร้างใหม่) | Only one config exists | [ ] |

### 4.3 Update Tier Settings - อัปเดตการตั้งค่า Tier (`PUT /api/config/tiers`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 4.3.1 | Update with valid tier array | อัปเดตด้วย array tier ที่ถูกต้อง | Tiers updated | [ ] |
| 4.3.2 | Update with empty tier array | อัปเดตด้วย array tier ว่าง | Handle gracefully | [ ] |
| 4.3.3 | Update with invalid tier structure | อัปเดตด้วยโครงสร้าง tier ไม่ถูกต้อง | Return error (400) | [ ] |
| 4.3.4 | Verify tier rewards updated | ตรวจสอบว่ารางวัล tier ถูกอัปเดต | Rewards match input | [ ] |
| 4.3.5 | Verify tier targets updated | ตรวจสอบว่าเป้าหมาย tier ถูกอัปเดต | Targets match input | [ ] |
| 4.3.6 | Verify tier levels updated | ตรวจสอบว่า levels ของ tier ถูกอัปเดต | Levels match input | [ ] |

### 4.4 Update Flex Messages - อัปเดตข้อความ Flex (`PUT /api/config/flex-messages`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 4.4.1 | Update all flex message templates | อัปเดต template ข้อความ flex ทั้งหมด | Templates updated | [ ] |
| 4.4.2 | Update single flex message | อัปเดต flex message เดียว | Only that message updated | [ ] |
| 4.4.3 | Update with invalid JSON structure | อัปเดตด้วยโครงสร้าง JSON ไม่ถูกต้อง | Return error (400) | [ ] |
| 4.4.4 | Verify placeholder syntax preserved | ตรวจสอบว่ารูปแบบ placeholder ยังอยู่ ({key}) | {key} not stripped | [ ] |

### 4.5 Update Site Template - อัปเดต Template เว็บไซต์ (`PUT /api/config/site-template`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 4.5.1 | Update site template config | อัปเดตการตั้งค่า template เว็บไซต์ | Template updated | [ ] |
| 4.5.2 | Update colors configuration | อัปเดตการตั้งค่าสี | Colors saved | [ ] |
| 4.5.3 | Update images configuration | อัปเดตการตั้งค่ารูปภาพ | Image URLs saved | [ ] |
| 4.5.4 | Update text content | อัปเดตเนื้อหาข้อความ | Text content saved | [ ] |

### 4.6 Upload Image - อัปโหลดรูปภาพ (`POST /api/config/upload-image`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 4.6.1 | Upload valid image (PNG) | อัปโหลดรูป PNG | Return public URL | [ ] |
| 4.6.2 | Upload valid image (JPG) | อัปโหลดรูป JPG | Return public URL | [ ] |
| 4.6.3 | Upload valid image (SVG) | อัปโหลดรูป SVG | Return public URL | [ ] |
| 4.6.4 | Upload with type "siteTemplate.logo" | อัปโหลดด้วย type โลโก้เว็บ | Correct path in storage | [ ] |
| 4.6.5 | Upload with type "flex_messages.success" | อัปโหลดด้วย type flex message | Correct path in storage | [ ] |
| 4.6.6 | Upload without image file | อัปโหลดโดยไม่มีไฟล์รูป | Return error (400) | [ ] |
| 4.6.7 | Upload without type parameter | อัปโหลดโดยไม่ระบุ type | Return error (400) | [ ] |
| 4.6.8 | Upload very large image | อัปโหลดรูปขนาดใหญ่มาก | Handle size limits | [ ] |
| 4.6.9 | Upload non-image file | อัปโหลดไฟล์ที่ไม่ใช่รูป | Return error (400) | [ ] |
| 4.6.10 | Verify public access on uploaded image | ตรวจสอบว่ารูปที่อัปโหลดเข้าถึงได้ | URL is accessible | [ ] |

---

## 5. Dashboard - แดชบอร์ด

### 5.1 Get Dashboard - ดึงข้อมูลแดชบอร์ด (`GET /api/dashboard`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 5.1.1 | Get dashboard with data | ดึงแดชบอร์ดเมื่อมีข้อมูล | Return total_missions, tier_stats | [ ] |
| 5.1.2 | Get dashboard when empty | ดึงแดชบอร์ดเมื่อไม่มีข้อมูล | Return zeros | [ ] |
| 5.1.3 | Verify tier statistics accuracy | ตรวจสอบความถูกต้องของสถิติ tier | Stats match actual data | [ ] |

### 5.2 Get Stats - ดึงสถิติ (`GET /api/dashboard/stats`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 5.2.1 | Get stats with data | ดึงสถิติเมื่อมีข้อมูล | Return stats array | [ ] |
| 5.2.2 | Verify total users count | ตรวจสอบจำนวนผู้ใช้ทั้งหมด | Count matches tbl_client | [ ] |
| 5.2.3 | Verify total missions count | ตรวจสอบจำนวนภารกิจทั้งหมด | Count matches tbl_mission | [ ] |
| 5.2.4 | Verify completed missions count | ตรวจสอบจำนวนภารกิจที่เสร็จสิ้น | Count is accurate | [ ] |
| 5.2.5 | Verify pending rewards count | ตรวจสอบจำนวนรางวัลรอดำเนินการ | Count is accurate | [ ] |

### 5.3 Get Tier Performance - ดึงประสิทธิภาพ Tier (`GET /api/dashboard/tier-performance`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 5.3.1 | Get tier performance data | ดึงข้อมูลประสิทธิภาพ tier | Return tier breakdown | [ ] |
| 5.3.2 | Verify active users per tier | ตรวจสอบผู้ใช้ที่ active ในแต่ละ tier | Counts accurate | [ ] |
| 5.3.3 | Verify completion rates | ตรวจสอบอัตราการผ่านสำเร็จ | Percentages accurate | [ ] |

### 5.4 Get Urgent Alerts - ดึงการแจ้งเตือนเร่งด่วน (`GET /api/dashboard/urgent-alerts`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 5.4.1 | Get alerts with expiring rewards | ดึงการแจ้งเตือนรางวัลใกล้หมดอายุ | Return reward alerts | [ ] |
| 5.4.2 | Get alerts with pending approvals | ดึงการแจ้งเตือนรอการอนุมัติ | Return approval alerts | [ ] |
| 5.4.3 | Get alerts with expiring levels | ดึงการแจ้งเตือน level ใกล้หมดอายุ | Return level alerts | [ ] |
| 5.4.4 | Get alerts when no urgent items | ดึงการแจ้งเตือนเมื่อไม่มีเรื่องเร่งด่วน | Return empty array | [ ] |

### 5.5 Get Recent Activities - ดึงกิจกรรมล่าสุด (`GET /api/dashboard/recent-activities`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 5.5.1 | Get recent activities | ดึงกิจกรรมล่าสุด | Return activities list | [ ] |
| 5.5.2 | Verify chronological order | ตรวจสอบลำดับเวลา (ใหม่สุดก่อน) | Newest first | [ ] |
| 5.5.3 | Verify activity types included | ตรวจสอบว่ามีทุกประเภทกิจกรรม | All types present | [ ] |
| 5.5.4 | Get activities when empty | ดึงกิจกรรมเมื่อไม่มีข้อมูล | Return empty array | [ ] |

### 5.6 Get Pending Rewards - ดึงรางวัลรอดำเนินการ (`GET /api/dashboard/pending-rewards`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 5.6.1 | Get pending rewards list | ดึงรายการรางวัลรอดำเนินการ | Return summary + list | [ ] |
| 5.6.2 | Verify pending count accuracy | ตรวจสอบความถูกต้องของจำนวนรอดำเนินการ | Count matches tbl_logs | [ ] |
| 5.6.3 | Get when no pending rewards | ดึงเมื่อไม่มีรางวัลรอดำเนินการ | Return empty list | [ ] |
| 5.6.4 | Verify reward details included | ตรวจสอบว่ารายละเอียดรางวัลครบ | All fields present | [ ] |

---

## 6. User Bet - ระบบยอดเดิมพันผู้ใช้

### 6.1 Get Current Bet - ดึงยอดเดิมพันปัจจุบัน (`GET /api/user-bet`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 6.1.1 | Get bet with valid userId and date range | ดึงยอดเดิมพันด้วย userId และช่วงวันที่ถูกต้อง | Return bet amount | [ ] |
| 6.1.2 | Get bet with missing userId | ดึงยอดเดิมพันโดยไม่มี userId | Return error (400) | [ ] |
| 6.1.3 | Get bet with missing date range | ดึงยอดเดิมพันโดยไม่ระบุช่วงวันที่ | Handle with defaults | [ ] |
| 6.1.4 | Get bet when external API fails | ดึงยอดเดิมพันเมื่อ API ภายนอกล้มเหลว | Return 0 | [ ] |
| 6.1.5 | Verify date range parsing | ตรวจสอบการแปลงช่วงวันที่ (Unix timestamp) | Correct dates sent to API | [ ] |
| 6.1.6 | Get bet for user with no betting data | ดึงยอดเดิมพันสำหรับผู้ใช้ที่ไม่มีข้อมูลเดิมพัน | Return 0 | [ ] |

### 6.2 Update Current Bet - อัปเดตยอดเดิมพัน (`PUT /api/user-bet`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 6.2.1 | Update bet with valid data | อัปเดตยอดเดิมพันด้วยข้อมูลถูกต้อง | Return success | [ ] |
| 6.2.2 | Update with missing userId | อัปเดตโดยไม่มี userId | Return error (400) | [ ] |
| 6.2.3 | Update with negative bet amount | อัปเดตด้วยยอดเดิมพันติดลบ | Handle gracefully | [ ] |
| 6.2.4 | Update with zero bet amount | อัปเดตด้วยยอดเดิมพัน 0 | Save successfully | [ ] |
| 6.2.5 | Verify UpdatedAt timestamp updated | ตรวจสอบว่า UpdatedAt ถูกอัปเดต | Timestamp is current | [ ] |

---

## 7. LINE Messaging - ระบบส่งข้อความ LINE

### 7.1 Send Follow-Up Flex Message - ส่งข้อความติดตามผล

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 7.1.1 | Send message with valid config | ส่งข้อความด้วย config ที่ถูกต้อง | Message sent to LINE | [ ] |
| 7.1.2 | Send message with invalid channel token | ส่งข้อความด้วย channel token ไม่ถูกต้อง | Handle error gracefully | [ ] |
| 7.1.3 | Send message with invalid user ID | ส่งข้อความด้วย user ID ไม่ถูกต้อง | Handle error gracefully | [ ] |
| 7.1.4 | Verify placeholders replaced | ตรวจสอบว่า placeholder ถูกแทนที่ (target, current_bet) | Values substituted | [ ] |
| 7.1.5 | Verify number formatting | ตรวจสอบการจัดรูปแบบตัวเลข (มี comma) | Numbers formatted | [ ] |
| 7.1.6 | Verify message logged | ตรวจสอบว่าข้อความถูกบันทึกใน tbl_logs_message | Log entry created | [ ] |

### 7.2 Send Mission Success Message - ส่งข้อความสำเร็จ

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 7.2.1 | Send success message | ส่งข้อความสำเร็จ | Message sent | [ ] |
| 7.2.2 | Verify level/tier info in message | ตรวจสอบข้อมูล level/tier ในข้อความ | Info is accurate | [ ] |
| 7.2.3 | Verify message logged | ตรวจสอบว่าข้อความถูกบันทึก | Log entry created | [ ] |

### 7.3 Send Mission Failed Message - ส่งข้อความล้มเหลว

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 7.3.1 | Send failed message | ส่งข้อความล้มเหลว | Message sent | [ ] |
| 7.3.2 | Verify failure reason in message | ตรวจสอบเหตุผลที่ล้มเหลวในข้อความ | Reason included | [ ] |
| 7.3.3 | Verify message logged | ตรวจสอบว่าข้อความถูกบันทึก | Log entry created | [ ] |

### 7.4 Send Mission Complete Message - ส่งข้อความจบภารกิจ

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 7.4.1 | Send complete message with reward expiry | ส่งข้อความจบภารกิจพร้อมวันหมดอายุรางวัล | Message sent | [ ] |
| 7.4.2 | Verify reward details in message | ตรวจสอบรายละเอียดรางวัลในข้อความ | Details accurate | [ ] |
| 7.4.3 | Verify expiry date in message | ตรวจสอบวันหมดอายุในข้อความ | Date correct | [ ] |
| 7.4.4 | Verify message logged | ตรวจสอบว่าข้อความถูกบันทึก | Log entry created | [ ] |

### 7.5 Send Get Reward Message - ส่งข้อความขอรับรางวัล

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 7.5.1 | Send reward claim notification | ส่งการแจ้งเตือนขอรับรางวัล | Message sent | [ ] |
| 7.5.2 | Verify reward amount in message | ตรวจสอบจำนวนรางวัลในข้อความ | Amount correct | [ ] |
| 7.5.3 | Verify message logged | ตรวจสอบว่าข้อความถูกบันทึก | Log entry created | [ ] |

### 7.6 Send Reward Notification Message - ส่งข้อความแจ้งเตือนรางวัล

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 7.6.1 | Send reward expiry reminder | ส่งการแจ้งเตือนรางวัลใกล้หมดอายุ | Message sent | [ ] |
| 7.6.2 | Verify remaining days in message | ตรวจสอบจำนวนวันที่เหลือในข้อความ | Days correct | [ ] |
| 7.6.3 | Verify message logged | ตรวจสอบว่าข้อความถูกบันทึก | Log entry created | [ ] |

---

## 8. Telegram Notification - ระบบแจ้งเตือน Telegram

### 8.1 Send Reward Claimed Message - ส่งข้อความขอรับรางวัลให้แอดมิน

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 8.1.1 | Send message with valid bot token | ส่งข้อความด้วย bot token ที่ถูกต้อง | Message sent | [ ] |
| 8.1.2 | Send message with invalid bot token | ส่งข้อความด้วย bot token ไม่ถูกต้อง | Handle error | [ ] |
| 8.1.3 | Send message with invalid chat ID | ส่งข้อความด้วย chat ID ไม่ถูกต้อง | Handle error | [ ] |
| 8.1.4 | Verify HTML formatting in message | ตรวจสอบรูปแบบ HTML ในข้อความ | Formatting correct | [ ] |
| 8.1.5 | Verify mission details in message | ตรวจสอบรายละเอียดภารกิจในข้อความ | Details accurate | [ ] |

---

## 9. Background Jobs - งานเบื้องหลัง (Event Processing)

### 9.1 Event Processing Loop - ลูปประมวลผล Event

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 9.1.1 | Process level_expiration event | ประมวลผล event หมดอายุ level | Level status updated | [ ] |
| 9.1.2 | Process follow_up event | ประมวลผล event ติดตามผล | Follow-up message sent | [ ] |
| 9.1.3 | Process reward_expiration event | ประมวลผล event รางวัลหมดอายุ | Expiration handled | [ ] |
| 9.1.4 | Process reward_notification event | ประมวลผล event แจ้งเตือนรางวัล | Notification sent | [ ] |
| 9.1.5 | Process recurring_reward_notification event | ประมวลผล event แจ้งเตือนรางวัลซ้ำ | Recurring notification sent | [ ] |
| 9.1.6 | Verify event status updated to "processed" | ตรวจสอบว่าสถานะ event เปลี่ยนเป็น "processed" | Status = processed | [ ] |
| 9.1.7 | Verify expired events are picked up | ตรวจสอบว่า event หมดอายุถูกหยิบขึ้นมาประมวลผล | All expired events processed | [ ] |
| 9.1.8 | Verify processing delay (100ms loop) | ตรวจสอบความถี่การประมวลผล (100ms) | Events processed promptly | [ ] |

### 9.2 Level Expiration Processing - การประมวลผล Level หมดอายุ

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 9.2.1 | Level expires with bet >= target | Level หมดอายุเมื่อยอดเดิมพัน >= เป้าหมาย | Level marked success | [ ] |
| 9.2.2 | Level expires with bet < target | Level หมดอายุเมื่อยอดเดิมพัน < เป้าหมาย | Level marked failed | [ ] |
| 9.2.3 | Level expires, consecutive fails reached max | Level หมดอายุ, ล้มเหลวติดต่อกันครบ | Mission failed | [ ] |
| 9.2.4 | Level expires, move to next level | Level หมดอายุ, ไปยัง level ถัดไป | New level created | [ ] |
| 9.2.5 | Level expires, tier completed | Level หมดอายุ, tier เสร็จสมบูรณ์ | Tier status = completed | [ ] |
| 9.2.6 | Verify external bet API called | ตรวจสอบว่า API ยอดเดิมพันถูกเรียก | API called with correct params | [ ] |
| 9.2.7 | Verify LINE messages sent | ตรวจสอบว่าส่งข้อความ LINE | Messages logged | [ ] |

### 9.3 Follow-Up Processing - การประมวลผลติดตามผล

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 9.3.1 | Follow-up event triggers message | Event ติดตามผลทำให้ส่งข้อความ | Message sent | [ ] |
| 9.3.2 | Verify current bet fetched | ตรวจสอบว่าดึงยอดเดิมพันปัจจุบัน | Bet amount in message | [ ] |
| 9.3.3 | Verify target shown | ตรวจสอบว่าแสดงเป้าหมาย | Target in message | [ ] |

---

## 10. External API Integration - การเชื่อมต่อ API ภายนอก

### 10.1 Phone Number Sync - ซิงค์เบอร์โทร (`POST /players/v1/line/sync`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 10.1.1 | Sync with valid phone number | ซิงค์ด้วยเบอร์โทรถูกต้อง | Return username | [ ] |
| 10.1.2 | Sync with invalid phone number | ซิงค์ด้วยเบอร์โทรไม่ถูกต้อง | Return error | [ ] |
| 10.1.3 | Sync with incorrect line_id | ซิงค์ด้วย line_id ไม่ตรง | Return error | [ ] |
| 10.1.4 | Verify API-KEY header sent | ตรวจสอบว่าส่ง API-KEY header | Header present | [ ] |
| 10.1.5 | Handle API timeout | จัดการเมื่อ API timeout | Return error gracefully | [ ] |
| 10.1.6 | Handle API 500 error | จัดการเมื่อ API ส่ง 500 | Return error gracefully | [ ] |

### 10.2 Get Bets - ดึงยอดเดิมพัน (`GET /players/v1/line/bets`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 10.2.1 | Get bet with valid params | ดึงยอดเดิมพันด้วยพารามิเตอร์ถูกต้อง | Return bet amount | [ ] |
| 10.2.2 | Get bet for user with no data | ดึงยอดเดิมพันสำหรับผู้ใช้ที่ไม่มีข้อมูล | Return 0 | [ ] |
| 10.2.3 | Verify date range in Unix timestamp | ตรวจสอบช่วงวันที่ในรูปแบบ Unix timestamp | Correct format | [ ] |
| 10.2.4 | Handle API timeout | จัดการเมื่อ API timeout | Return 0 | [ ] |
| 10.2.5 | Handle API error | จัดการเมื่อ API error | Return 0 | [ ] |
| 10.2.6 | Verify API-KEY header sent | ตรวจสอบว่าส่ง API-KEY header | Header present | [ ] |

### 10.3 Claim Reward - ส่งคำขอรับรางวัล (`POST /players/v1/line/rewards/claim`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 10.3.1 | Submit claim with valid data | ส่งคำขอรับรางวัลด้วยข้อมูลถูกต้อง | Claim submitted | [ ] |
| 10.3.2 | Verify callback URL sent | ตรวจสอบว่าส่ง callback URL | URL is correct | [ ] |
| 10.3.3 | Verify log_id sent | ตรวจสอบว่าส่ง log_id | ID is correct | [ ] |
| 10.3.4 | Verify mission details sent | ตรวจสอบว่าส่งรายละเอียดภารกิจ | Details accurate | [ ] |
| 10.3.5 | Handle API error | จัดการเมื่อ API error | Error handled gracefully | [ ] |

---

## 11. File Upload - ระบบอัปโหลดไฟล์ (Firebase)

### 11.1 Firebase Storage Upload - อัปโหลดไปยัง Firebase Storage

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 11.1.1 | Upload PNG image | อัปโหลดรูป PNG | Upload successful | [ ] |
| 11.1.2 | Upload JPG image | อัปโหลดรูป JPG | Upload successful | [ ] |
| 11.1.3 | Upload SVG image | อัปโหลดรูป SVG พร้อม content-type ถูกต้อง | Upload successful with correct content-type | [ ] |
| 11.1.4 | Verify public ACL set | ตรวจสอบว่าตั้งค่า public ACL | File publicly accessible | [ ] |
| 11.1.5 | Verify correct path structure | ตรวจสอบโครงสร้าง path ถูกต้อง | Path matches type param | [ ] |
| 11.1.6 | Verify MediaLink returned | ตรวจสอบว่าส่งคืน MediaLink | Valid public URL | [ ] |
| 11.1.7 | Upload with invalid Firebase config | อัปโหลดด้วย Firebase config ไม่ถูกต้อง | Error handled | [ ] |
| 11.1.8 | Upload very large file | อัปโหลดไฟล์ขนาดใหญ่มาก | Size limit enforced | [ ] |

---

## 12. Generic CRUD Operations - การดำเนินการ CRUD ทั่วไป

### 12.1 Create - สร้างข้อมูล (`POST /api/{collection}`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 12.1.1 | Create document with valid data | สร้าง document ด้วยข้อมูลถูกต้อง | Return ID (201) | [ ] |
| 12.1.2 | Create with empty body | สร้างด้วย body ว่าง | Return error (400) | [ ] |
| 12.1.3 | Create with invalid JSON | สร้างด้วย JSON ไม่ถูกต้อง | Return error (400) | [ ] |
| 12.1.4 | Verify _id generated | ตรวจสอบว่า _id ถูกสร้าง | ObjectID present | [ ] |

### 12.2 Get All - ดึงข้อมูลทั้งหมด (`GET /api/{collection}`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 12.2.1 | Get all with pagination | ดึงทั้งหมดพร้อม pagination | Return paginated results | [ ] |
| 12.2.2 | Get all when empty | ดึงทั้งหมดเมื่อไม่มีข้อมูล | Return empty array | [ ] |
| 12.2.3 | Verify sort order (descending) | ตรวจสอบลำดับการเรียง (ใหม่สุดก่อน) | Newest first | [ ] |
| 12.2.4 | Verify pagination metadata | ตรวจสอบข้อมูล pagination | Metadata accurate | [ ] |

### 12.3 Search - ค้นหา (`GET /api/{collection}/search`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 12.3.1 | Search with matching query | ค้นหาด้วยคำค้นที่ตรง | Return matching results | [ ] |
| 12.3.2 | Search with no matches | ค้นหาด้วยคำค้นที่ไม่ตรง | Return empty array | [ ] |
| 12.3.3 | Search with special regex characters | ค้นหาด้วยอักขระพิเศษ regex | Handle safely | [ ] |
| 12.3.4 | Search with field parameter | ค้นหาใน field เฉพาะ | Search specific field | [ ] |
| 12.3.5 | Verify case-insensitive search | ตรวจสอบการค้นหาไม่สนใจตัวพิมพ์ | Match regardless of case | [ ] |

### 12.4 Get by ID - ดึงตาม ID (`GET /api/{collection}/:id`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 12.4.1 | Get with valid ID | ดึงด้วย ID ที่ถูกต้อง | Return document | [ ] |
| 12.4.2 | Get with invalid ID format | ดึงด้วย ID รูปแบบไม่ถูกต้อง | Return error (400) | [ ] |
| 12.4.3 | Get with non-existent ID | ดึงด้วย ID ที่ไม่มีในระบบ | Return error (404) | [ ] |

### 12.5 Update - อัปเดต (`PUT /api/{collection}/:id`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 12.5.1 | Update with valid data | อัปเดตด้วยข้อมูลถูกต้อง | Return updated document | [ ] |
| 12.5.2 | Update with empty body | อัปเดตด้วย body ว่าง | Handle gracefully | [ ] |
| 12.5.3 | Update non-existent ID | อัปเดต ID ที่ไม่มีในระบบ | Return error (404) | [ ] |
| 12.5.4 | Partial update (some fields) | อัปเดตบาง field | Only specified fields updated | [ ] |

### 12.6 Delete - ลบ (`DELETE /api/{collection}/:id`)

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 12.6.1 | Delete existing document | ลบ document ที่มีอยู่ | Return success | [ ] |
| 12.6.2 | Delete non-existent document | ลบ document ที่ไม่มี | Return error (404) | [ ] |
| 12.6.3 | Delete with invalid ID | ลบด้วย ID ไม่ถูกต้อง | Return error (400) | [ ] |

---

## 13. Security Testing - การทดสอบความปลอดภัย

### 13.1 Authentication & Authorization - การยืนยันตัวตนและการอนุญาต

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 13.1.1 | Access protected endpoint without token | เข้าถึง endpoint ที่ต้องมี token โดยไม่มี token | Return 401 | [ ] |
| 13.1.2 | Access with expired token | เข้าถึงด้วย token หมดอายุ | Return 401 | [ ] |
| 13.1.3 | Access with invalid token signature | เข้าถึงด้วย token ที่ลายเซ็นไม่ถูกต้อง | Return 401 | [ ] |
| 13.1.4 | Access with manipulated claims | เข้าถึงด้วย claims ที่ถูกแก้ไข | Return 401 | [ ] |
| 13.1.5 | Verify JWT secret not exposed | ตรวจสอบว่า JWT secret ไม่ถูกเปิดเผย | Secret not in responses | [ ] |

### 13.2 Input Validation & Injection - การตรวจสอบ Input และการ Injection

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 13.2.1 | SQL injection in search query | ทดสอบ SQL injection ในการค้นหา | No database leak | [ ] |
| 13.2.2 | NoSQL injection in search query | ทดสอบ NoSQL injection ในการค้นหา | No database leak | [ ] |
| 13.2.3 | XSS in input fields | ทดสอบ XSS ใน input fields | Sanitized/escaped | [ ] |
| 13.2.4 | Command injection in file names | ทดสอบ command injection ในชื่อไฟล์ | Blocked | [ ] |
| 13.2.5 | Path traversal in file upload | ทดสอบ path traversal ในการอัปโหลด | Blocked | [ ] |
| 13.2.6 | BSON injection in ObjectID | ทดสอบ BSON injection ใน ObjectID | Rejected | [ ] |

### 13.3 Sensitive Data - ข้อมูลละเอียดอ่อน

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 13.3.1 | Password not in API responses | รหัสผ่านไม่ถูกส่งกลับใน API | Password hidden | [ ] |
| 13.3.2 | API keys not exposed | API keys ไม่ถูกเปิดเผย | Keys hidden/masked | [ ] |
| 13.3.3 | Firebase credentials protected | Firebase credentials ถูกปกป้อง | Not in responses | [ ] |
| 13.3.4 | LINE channel secret protected | LINE channel secret ถูกปกป้อง | Not in responses | [ ] |
| 13.3.5 | Telegram bot token protected | Telegram bot token ถูกปกป้อง | Not in responses | [ ] |

### 13.4 Rate Limiting - การจำกัดอัตราการเรียก

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 13.4.1 | Rapid login attempts | พยายาม login อย่างรวดเร็วหลายครั้ง | Rate limited | [ ] |
| 13.4.2 | Rapid API calls | เรียก API อย่างรวดเร็วหลายครั้ง | Rate limited | [ ] |
| 13.4.3 | Brute force password | ทดสอบ brute force รหัสผ่าน | Account locked/rate limited | [ ] |

### 13.5 CORS & Headers - CORS และ Headers

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 13.5.1 | CORS headers present | มี CORS headers | Correct origins allowed | [ ] |
| 13.5.2 | Security headers present | มี security headers (X-Frame-Options, etc.) | Headers present | [ ] |
| 13.5.3 | No sensitive data in headers | ไม่มีข้อมูลละเอียดอ่อนใน headers | Headers clean | [ ] |

---

## 14. Performance Testing - การทดสอบประสิทธิภาพ

### 14.1 Load Testing - การทดสอบโหลด

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 14.1.1 | 100 concurrent requests | 100 requests พร้อมกัน | Response < 1s | [ ] |
| 14.1.2 | 500 concurrent requests | 500 requests พร้อมกัน | Response < 3s | [ ] |
| 14.1.3 | 1000 concurrent requests | 1000 requests พร้อมกัน | No crashes | [ ] |
| 14.1.4 | Sustained load (10 min) | โหลดต่อเนื่อง 10 นาที | No memory leaks | [ ] |

### 14.2 Database Performance - ประสิทธิภาพฐานข้อมูล

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 14.2.1 | Query with 10,000 missions | Query กับ 10,000 missions | Response < 2s | [ ] |
| 14.2.2 | Query with 100,000 clients | Query กับ 100,000 clients | Response < 5s | [ ] |
| 14.2.3 | Pagination with large dataset | Pagination กับข้อมูลขนาดใหญ่ | Efficient query | [ ] |
| 14.2.4 | Search with large dataset | ค้นหากับข้อมูลขนาดใหญ่ | Indexed search | [ ] |

### 14.3 Event Processing Performance - ประสิทธิภาพการประมวลผล Event

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 14.3.1 | Process 100 pending events | ประมวลผล 100 events ที่รอ | Complete < 1s | [ ] |
| 14.3.2 | Process 1000 pending events | ประมวลผล 1000 events ที่รอ | Complete < 10s | [ ] |
| 14.3.3 | No event backlog under load | ไม่มี event ค้างภายใต้โหลด | Events processed timely | [ ] |

### 14.4 External API Performance - ประสิทธิภาพ API ภายนอก

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 14.4.1 | LINE API response time | เวลาตอบสนอง LINE API | < 2s | [ ] |
| 14.4.2 | External Players API response | เวลาตอบสนอง External Players API | < 2s | [ ] |
| 14.4.3 | Firebase upload response | เวลาตอบสนอง Firebase upload | < 5s | [ ] |
| 14.4.4 | Handle API timeout gracefully | จัดการ API timeout อย่างนุ่มนวล | No blocking | [ ] |

---

## 15. Database & Data Integrity - ฐานข้อมูลและความถูกต้องของข้อมูล

### 15.1 Data Consistency - ความสอดคล้องของข้อมูล

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 15.1.1 | Mission tier matches config | Tier ของ mission ตรงกับ config | Tier data consistent | [ ] |
| 15.1.2 | Level dates calculated correctly | วันที่ level ถูกคำนวณถูกต้อง | Dates accurate | [ ] |
| 15.1.3 | Event times match level expiry | เวลา event ตรงกับวันหมดอายุ level | Times synchronized | [ ] |
| 15.1.4 | Log mission_id matches mission | mission_id ใน log ตรงกับ mission | IDs consistent | [ ] |
| 15.1.5 | Client user_id matches mission | user_id ของ client ตรงกับ mission | IDs consistent | [ ] |

### 15.2 Concurrent Operations - การดำเนินการพร้อมกัน

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 15.2.1 | Concurrent mission updates | อัปเดต mission พร้อมกัน | No race conditions | [ ] |
| 15.2.2 | Concurrent reward claims | ขอรางวัลพร้อมกัน | Only one succeeds | [ ] |
| 15.2.3 | Concurrent event processing | ประมวลผล event พร้อมกัน | No duplicate processing | [ ] |
| 15.2.4 | Concurrent config updates | อัปเดต config พร้อมกัน | Last write wins | [ ] |

### 15.3 Orphan Data - ข้อมูลกำพร้า

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 15.3.1 | Delete client with missions | ลบ client ที่มี missions | Missions handled | [ ] |
| 15.3.2 | Delete mission with events | ลบ mission ที่มี events | Events cleaned up | [ ] |
| 15.3.3 | Delete mission with logs | ลบ mission ที่มี logs | Logs handled | [ ] |

---

## 16. Error Handling - การจัดการข้อผิดพลาด

### 16.1 API Error Responses - การตอบกลับข้อผิดพลาด API

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 16.1.1 | 400 errors have clear message | Error 400 มีข้อความชัดเจน | Message descriptive | [ ] |
| 16.1.2 | 404 errors have clear message | Error 404 มีข้อความชัดเจน | Resource identified | [ ] |
| 16.1.3 | 500 errors don't leak internals | Error 500 ไม่เปิดเผยข้อมูลภายใน | Generic message | [ ] |
| 16.1.4 | Error response format consistent | รูปแบบ error response สม่ำเสมอ | Same structure | [ ] |

### 16.2 External Service Failures - ความล้มเหลวของบริการภายนอก

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 16.2.1 | LINE API unavailable | LINE API ไม่พร้อมใช้งาน | Graceful degradation | [ ] |
| 16.2.2 | External Players API down | External Players API ล่ม | Fallback behavior | [ ] |
| 16.2.3 | Firebase unavailable | Firebase ไม่พร้อมใช้งาน | Error message returned | [ ] |
| 16.2.4 | MongoDB connection lost | สูญเสียการเชื่อมต่อ MongoDB | Auto-reconnect | [ ] |
| 16.2.5 | Telegram API failure | Telegram API ล้มเหลว | Mission continues | [ ] |

### 16.3 Invalid Data Handling - การจัดการข้อมูลไม่ถูกต้อง

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 16.3.1 | Invalid ObjectID format | รูปแบบ ObjectID ไม่ถูกต้อง | Clear error message | [ ] |
| 16.3.2 | Invalid JSON body | JSON body ไม่ถูกต้อง | Clear error message | [ ] |
| 16.3.3 | Missing required fields | ขาด field ที่จำเป็น | Field names in error | [ ] |
| 16.3.4 | Invalid date format | รูปแบบวันที่ไม่ถูกต้อง | Clear error message | [ ] |
| 16.3.5 | Invalid enum values | ค่า enum ไม่ถูกต้อง | Clear error message | [ ] |

---

## 17. Edge Cases & Boundary Testing - กรณีขอบเขตและการทดสอบขีดจำกัด

### 17.1 Mission Edge Cases - กรณีขอบเขตของภารกิจ

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 17.1.1 | Complete all 3 tiers | ผ่านครบทั้ง 3 tiers | Mission fully completed | [ ] |
| 17.1.2 | Fail at max consecutive fails | ล้มเหลวติดต่อกันครบจำนวนสูงสุด | Mission failed | [ ] |
| 17.1.3 | Claim reward at exact expiry time | ขอรางวัลตอนหมดอายุพอดี | Handle boundary | [ ] |
| 17.1.4 | Level expires at midnight | Level หมดอายุตอนเที่ยงคืน | Time zone handled | [ ] |
| 17.1.5 | User starts new mission after complete | ผู้ใช้เริ่มภารกิจใหม่หลังจบ | New mission created | [ ] |
| 17.1.6 | User starts new mission after fail | ผู้ใช้เริ่มภารกิจใหม่หลังล้มเหลว | New mission created | [ ] |
| 17.1.7 | Bet exactly equals target | ยอดเดิมพันเท่ากับเป้าหมายพอดี | Count as success | [ ] |
| 17.1.8 | Bet is 0 | ยอดเดิมพันเป็น 0 | Count as failure | [ ] |
| 17.1.9 | Bet is negative (if possible) | ยอดเดิมพันติดลบ (ถ้าเป็นไปได้) | Handle gracefully | [ ] |

### 17.2 Pagination Edge Cases - กรณีขอบเขตของ Pagination

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 17.2.1 | Page 0 request | ขอหน้า 0 | Default to page 1 | [ ] |
| 17.2.2 | Page beyond total pages | ขอหน้าเกินจำนวนหน้าทั้งหมด | Empty array | [ ] |
| 17.2.3 | Limit 0 request | ขอ limit 0 | Default limit used | [ ] |
| 17.2.4 | Limit negative request | ขอ limit ติดลบ | Default limit used | [ ] |
| 17.2.5 | Exactly one page of results | ผลลัพธ์พอดี 1 หน้า | has_next = false | [ ] |
| 17.2.6 | Last page of results | หน้าสุดท้ายของผลลัพธ์ | has_next = false | [ ] |

### 17.3 Time/Date Edge Cases - กรณีขอบเขตของเวลา/วันที่

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 17.3.1 | Daylight saving time transition | การเปลี่ยนเวลาออมแสง | Dates handled | [ ] |
| 17.3.2 | Leap year date handling | การจัดการวันที่ปีอธิกสุรทิน | Dates valid | [ ] |
| 17.3.3 | Year boundary (Dec 31 - Jan 1) | ขอบเขตปี (31 ธ.ค. - 1 ม.ค.) | Dates correct | [ ] |
| 17.3.4 | UTC vs local time handling | การจัดการ UTC vs เวลาท้องถิ่น | Consistent timezone | [ ] |

### 17.4 String Edge Cases - กรณีขอบเขตของข้อความ

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 17.4.1 | Empty string in required fields | ข้อความว่างใน field ที่จำเป็น | Rejected | [ ] |
| 17.4.2 | Very long string (10000+ chars) | ข้อความยาวมาก (10000+ ตัวอักษร) | Handled/truncated | [ ] |
| 17.4.3 | Unicode characters (Thai, Emoji) | อักขระ Unicode (ไทย, Emoji) | Stored correctly | [ ] |
| 17.4.4 | Special characters (!@#$%^&*) | อักขระพิเศษ (!@#$%^&*) | Escaped/handled | [ ] |
| 17.4.5 | Null byte in string | Null byte ในข้อความ | Rejected/handled | [ ] |

### 17.5 Numeric Edge Cases - กรณีขอบเขตของตัวเลข

| # | Test Case | คำอธิบาย | Expected Result | Status |
|---|-----------|----------|-----------------|--------|
| 17.5.1 | Bet amount with decimals | ยอดเดิมพันที่มีทศนิยม | Handled correctly | [ ] |
| 17.5.2 | Very large bet amount | ยอดเดิมพันขนาดใหญ่มาก | No overflow | [ ] |
| 17.5.3 | Target of 0 | เป้าหมายเป็น 0 | Handle edge case | [ ] |
| 17.5.4 | Reward amount with decimals | รางวัลที่มีทศนิยม | Formatted correctly | [ ] |

---

## สรุปการทดสอบ (Test Summary)

### รวมทั้งหมด (Total Test Cases)

| Section | หมวด | จำนวน Test Cases |
|---------|------|-----------------|
| 1. Admin Authentication | ระบบยืนยันตัวตนผู้ดูแล | 17 |
| 2. Client Management | ระบบจัดการลูกค้า | 29 |
| 3. Mission Management | ระบบจัดการภารกิจ | 46 |
| 4. Configuration Management | ระบบจัดการการตั้งค่า | 29 |
| 5. Dashboard | แดชบอร์ด | 18 |
| 6. User Bet | ระบบยอดเดิมพันผู้ใช้ | 11 |
| 7. LINE Messaging | ระบบส่งข้อความ LINE | 18 |
| 8. Telegram Notification | ระบบแจ้งเตือน Telegram | 5 |
| 9. Background Jobs | งานเบื้องหลัง | 17 |
| 10. External API Integration | การเชื่อมต่อ API ภายนอก | 17 |
| 11. File Upload | ระบบอัปโหลดไฟล์ | 8 |
| 12. Generic CRUD Operations | การดำเนินการ CRUD ทั่วไป | 18 |
| 13. Security Testing | การทดสอบความปลอดภัย | 18 |
| 14. Performance Testing | การทดสอบประสิทธิภาพ | 14 |
| 15. Database & Data Integrity | ฐานข้อมูลและความถูกต้อง | 12 |
| 16. Error Handling | การจัดการข้อผิดพลาด | 14 |
| 17. Edge Cases & Boundary | กรณีขอบเขต | 27 |
| **รวมทั้งหมด (Total)** | | **318** |

---

## เครื่องมือที่แนะนำสำหรับการทดสอบ (Recommended Testing Tools)

### API Testing - ทดสอบ API
- **Postman** - ทดสอบ API แบบ manual
- **Bruno** - API client น้ำหนักเบา
- **k6** - ทดสอบโหลด
- **Artillery** - ทดสอบประสิทธิภาพ

### Database Testing - ทดสอบฐานข้อมูล
- **MongoDB Compass** - ตรวจสอบข้อมูล
- **Mongosh** - รัน query ฐานข้อมูล

### Security Testing - ทดสอบความปลอดภัย
- **OWASP ZAP** - สแกนช่องโหว่
- **Burp Suite** - ทดสอบความปลอดภัย
- **sqlmap** - ทดสอบ injection

### Monitoring - ติดตามระบบ
- **Prometheus** - เก็บ metrics
- **Grafana** - แสดงผล
- **Jaeger** - distributed tracing

---

## ลำดับความสำคัญในการทดสอบ (Testing Priority)

### Priority 1 - สำคัญมาก (Critical)
1. **Mission Management** - เป็น core feature หลัก
2. **Reward Callback** - เกี่ยวข้องกับเงิน
3. **Authentication** - ความปลอดภัยของระบบ

### Priority 2 - สำคัญ (High)
1. **Background Jobs** - ต้องทำงานถูกต้องตลอดเวลา
2. **External API Integration** - เชื่อมต่อระบบภายนอก
3. **LINE Messaging** - แจ้งเตือนผู้ใช้

### Priority 3 - ปานกลาง (Medium)
1. **Client Management** - จัดการข้อมูลลูกค้า
2. **Configuration** - ตั้งค่าระบบ
3. **Dashboard** - แสดงผลข้อมูล

### Priority 4 - ต่ำ (Low)
1. **Generic CRUD** - การดำเนินการพื้นฐาน
2. **File Upload** - อัปโหลดรูปภาพ

---

## Critical Path - เส้นทางวิกฤต

```
สร้างภารกิจ (Create Mission)
    ↓
อัปเดตสถานะ (Update Status) - ผ่าน/ไม่ผ่าน
    ↓
จบ Tier (Complete Tier)
    ↓
ขอรับรางวัล (Claim Reward)
    ↓
รอการอนุมัติ (Pending Approval)
    ↓
Callback อนุมัติ/ปฏิเสธ (Approve/Reject)
    ↓
ไป Tier ถัดไป หรือ จบภารกิจ
```

**ต้องทดสอบ Critical Path นี้ให้ครบถ้วนก่อนนำขึ้น Production**

---

## หมายเหตุ (Notes)

1. **Priority Testing:** เริ่มจาก Mission Management เนื่องจากเป็น core feature
2. **Critical Path:** Create Mission → Update Status → Claim Reward → Callback
3. **External Dependencies:** ทดสอบ mock API ก่อน แล้วค่อยทดสอบกับ production API
4. **Background Jobs:** ต้องทดสอบ timing อย่างละเอียด เนื่องจากมีผลต่อ user experience
5. **Security:** ต้องทดสอบก่อน deploy to production

---

*Document generated: 2025-12-08*
*เอกสารสร้างโดย: Claude Code*
