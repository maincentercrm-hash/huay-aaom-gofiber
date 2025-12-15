/**
 * Mock Data สำหรับ Dashboard Components
 * จำลองข้อมูลจาก Models ตาม model_data.md
 */

// KPI Stats Cards
export const mockStatsData = [
  {
    title: 'ผู้ใช้ทั้งหมด',
    value: 1234,
    icon: 'tabler-users',
    color: 'primary' as const,
    trend: {
      value: 12,
      label: 'จากเดือนที่แล้ว'
    }
  },
  {
    title: 'มิชชันกำลังทำ',
    value: 456,
    icon: 'tabler-rocket',
    color: 'info' as const,
    trend: {
      value: 8,
      label: 'จากสัปดาห์ที่แล้ว'
    }
  },
  {
    title: 'สำเร็จแล้ว',
    value: 789,
    icon: 'tabler-circle-check',
    color: 'success' as const,
    trend: {
      value: 15,
      label: 'จากสัปดาห์ที่แล้ว'
    }
  },
  {
    title: 'รางวัลรอแจก',
    value: 45,
    icon: 'tabler-gift',
    color: 'warning' as const,
    trend: {
      value: -5,
      label: 'จากเมื่อวาน'
    }
  }
]

// Tier Performance Data
export const mockTierPerformanceData = [
  {
    tier: 1,
    name: 'TIER 1',
    totalUsers: 900,
    completedUsers: 702,
    processingUsers: 150,
    failedUsers: 48,
    successRate: 78,
    color: '#22c55e', // green
    activeUsers: [
      {
        id: '1',
        userId: 'U1234567890',
        displayName: 'สมชาย ใจดี',
        pictureUrl: 'https://i.pravatar.cc/150?img=12',
        phoneNumber: '089-123-4567',
        currentLevel: 3,
        status: 'processing',
        updatedAt: '5 นาทีที่แล้ว'
      },
      {
        id: '2',
        userId: 'U0987654321',
        displayName: 'สมหญิง รักสวย',
        pictureUrl: 'https://i.pravatar.cc/150?img=23',
        phoneNumber: '081-987-6543',
        currentLevel: 2,
        status: 'processing',
        updatedAt: '12 นาทีที่แล้ว'
      },
      {
        id: '3',
        userId: 'U1122334455',
        displayName: 'วิชัย มั่นคง',
        pictureUrl: 'https://i.pravatar.cc/150?img=33',
        phoneNumber: '092-345-6789',
        currentLevel: 4,
        status: 'processing',
        updatedAt: '25 นาทีที่แล้ว'
      },
      {
        id: '4',
        userId: 'U5566778899',
        displayName: 'มาลี สวยงาม',
        pictureUrl: 'https://i.pravatar.cc/150?img=44',
        phoneNumber: '086-789-0123',
        currentLevel: 1,
        status: 'processing',
        updatedAt: '1 ชั่วโมงที่แล้ว'
      }
    ]
  },
  {
    tier: 2,
    name: 'TIER 2',
    totalUsers: 400,
    completedUsers: 284,
    processingUsers: 80,
    failedUsers: 36,
    successRate: 71,
    color: '#3b82f6', // blue
    activeUsers: [
      {
        id: '5',
        userId: 'U2233445566',
        displayName: 'ประยุทธ์ ชนะเลิศ',
        pictureUrl: 'https://i.pravatar.cc/150?img=55',
        phoneNumber: '095-234-5678',
        currentLevel: 2,
        status: 'processing',
        updatedAt: '8 นาทีที่แล้ว'
      },
      {
        id: '6',
        userId: 'U6677889900',
        displayName: 'วารี น้ำใส',
        pictureUrl: 'https://i.pravatar.cc/150?img=26',
        phoneNumber: '088-456-7890',
        currentLevel: 3,
        status: 'processing',
        updatedAt: '18 นาทีที่แล้ว'
      },
      {
        id: '7',
        userId: 'U3344556677',
        displayName: 'สุรชัย ดีมาก',
        pictureUrl: 'https://i.pravatar.cc/150?img=67',
        phoneNumber: '091-567-8901',
        currentLevel: 1,
        status: 'processing',
        updatedAt: '42 นาทีที่แล้ว'
      }
    ]
  },
  {
    tier: 3,
    name: 'TIER 3',
    totalUsers: 165,
    completedUsers: 122,
    processingUsers: 28,
    failedUsers: 15,
    successRate: 74,
    color: '#a855f7', // purple
    activeUsers: [
      {
        id: '8',
        userId: 'U7788990011',
        displayName: 'ธนากร เศรษฐี',
        pictureUrl: 'https://i.pravatar.cc/150?img=8',
        phoneNumber: '087-678-9012',
        currentLevel: 2,
        status: 'processing',
        updatedAt: '15 นาทีที่แล้ว'
      },
      {
        id: '9',
        userId: 'U4455667788',
        displayName: 'นิภา สุขสันต์',
        pictureUrl: 'https://i.pravatar.cc/150?img=29',
        phoneNumber: '093-789-0123',
        currentLevel: 1,
        status: 'processing',
        updatedAt: '32 นาทีที่แล้ว'
      }
    ]
  }
]

// Urgent Alerts Data
export const mockUrgentAlertsData = [
  {
    id: '1',
    type: 'reward_expiring' as const,
    title: 'รางวัลใกล้หมดอายุ',
    description: 'มีรางวัลที่จะหมดอายุภายใน 24 ชั่วโมง',
    count: 5,
    severity: 'error' as const,
    icon: 'tabler-alarm',
    actionLabel: 'ดูรายละเอียด'
  },
  {
    id: '2',
    type: 'approval_pending' as const,
    title: 'รอการอนุมัติ',
    description: 'มีรางวัลรอการอนุมัติจากระบบ',
    count: 12,
    severity: 'warning' as const,
    icon: 'tabler-hourglass',
    actionLabel: 'ดำเนินการ'
  },
  {
    id: '3',
    type: 'level_expiring' as const,
    title: 'Level หมดอายุวันนี้',
    description: 'มี Level ที่จะหมดอายุวันนี้',
    count: 8,
    severity: 'info' as const,
    icon: 'tabler-calendar-event',
    actionLabel: 'ดูรายชื่อ'
  }
]

// Reward Management Data
export const mockRewardManagementData = {
  pendingApproval: 45,
  pendingAmount: 234500,
  approvedCount: 890,
  approvedAmount: 4567800,
  rejectedCount: 23,
  totalAmount: 4802300,
  approvalRate: 97.5
}

// Recent Activities Data
export const mockRecentActivitiesData = [
  {
    id: '1',
    type: 'mission_complete' as const,
    userPhone: '089-xxx-1234',
    description: 'ทำ Tier 2 สำเร็จ',
    timestamp: '2 นาที',
    tier: 2,
    level: 5,
    reward: 5000
  },
  {
    id: '2',
    type: 'reward_claim' as const,
    userPhone: '081-xxx-5678',
    description: 'รับรางวัลสำเร็จ',
    timestamp: '5 นาที',
    reward: 3000
  },
  {
    id: '3',
    type: 'level_complete' as const,
    userPhone: '092-xxx-9012',
    description: 'ผ่าน Level 3 ของ Tier 1',
    timestamp: '12 นาที',
    tier: 1,
    level: 3
  },
  {
    id: '4',
    type: 'new_user' as const,
    userPhone: '086-xxx-3456',
    description: 'ผู้ใช้ใหม่ลงทะเบียน',
    timestamp: '18 นาที'
  },
  {
    id: '5',
    type: 'mission_fail' as const,
    userPhone: '095-xxx-7890',
    description: 'Tier 1 Level 2 หมดเวลา',
    timestamp: '25 นาที',
    tier: 1,
    level: 2
  },
  {
    id: '6',
    type: 'level_complete' as const,
    userPhone: '088-xxx-2345',
    description: 'ผ่าน Level 1 ของ Tier 3',
    timestamp: '32 นาที',
    tier: 3,
    level: 1
  },
  {
    id: '7',
    type: 'reward_claim' as const,
    userPhone: '091-xxx-6789',
    description: 'รับรางวัลสำเร็จ',
    timestamp: '45 นาที',
    reward: 8000
  },
  {
    id: '8',
    type: 'new_user' as const,
    userPhone: '087-xxx-0123',
    description: 'ผู้ใช้ใหม่ลงทะเบียน',
    timestamp: '1 ชั่วโมง'
  }
]
