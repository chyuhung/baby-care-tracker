package models

import "time"

// Family 家庭组
type Family struct {
	ID         int64     `json:"id"`
	InviteCode string    `json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`
}

// FamilyMember 家庭成员
type FamilyMember struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// User 用户
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	FamilyID     *int64    `json:"family_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// JoinFamilyRequest 加入家庭请求
type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// Baby 宝宝
type Baby struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	BirthDate   string    `json:"birth_date"`
	Gender      string    `json:"gender"`
	AvatarColor string    `json:"avatar_color"`
	CreatedAt   time.Time `json:"created_at"`
}

// FeedingRecord 喂奶记录
type FeedingRecord struct {
	ID              int64  `json:"id"`
	BabyID          int64  `json:"baby_id"`
	UserID          int64  `json:"user_id"`
	Type            string `json:"type"` // breast, bottle, formula
	DurationMinutes int    `json:"duration_minutes"`
	AmountMl        int    `json:"amount_ml"`
	Side            string `json:"side"` // left, right, both
	Brand           string `json:"brand"`
	Note            string `json:"note"`
	OccurredAt      string `json:"occurred_at"`
	CreatedAt       string `json:"created_at"`
	RecordType      string `json:"record_type"` // feeding
}

// DiaperRecord 尿布记录
type DiaperRecord struct {
	ID         int64  `json:"id"`
	BabyID     int64  `json:"baby_id"`
	UserID     int64  `json:"user_id"`
	Type       string `json:"type"` // pee, poop, mixed
	Note       string `json:"note"`
	OccurredAt string `json:"occurred_at"`
	CreatedAt  string `json:"created_at"`
	RecordType string `json:"record_type"` // diaper
}

// Record 统一记录类型
type Record struct {
	ID         int64  `json:"id"`
	BabyID     int64  `json:"baby_id"`
	UserID     int64  `json:"user_id"`
	RecordType string `json:"record_type"` // feeding, diaper
	Data       any    `json:"data"`
	OccurredAt string `json:"occurred_at"`
	CreatedAt  string `json:"created_at"`
}

// --- Request / Response DTOs ---

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateBabyRequest struct {
	Name        string `json:"name" binding:"required"`
	BirthDate   string `json:"birth_date" binding:"required"`
	Gender      string `json:"gender"`
	AvatarColor string `json:"avatar_color"`
}

type UpdateBabyRequest struct {
	Name        string `json:"name"`
	BirthDate   string `json:"birth_date"`
	Gender      string `json:"gender"`
	AvatarColor string `json:"avatar_color"`
}

type CreateFeedingRequest struct {
	Type            string `json:"type" binding:"required"`
	DurationMinutes int    `json:"duration_minutes"`
	AmountMl        int    `json:"amount_ml"`
	Side            string `json:"side"`
	Brand           string `json:"brand"`
	Note            string `json:"note"`
	OccurredAt      string `json:"occurred_at" binding:"required"`
}

type CreateDiaperRequest struct {
	Type       string `json:"type" binding:"required"`
	Note       string `json:"note"`
	OccurredAt string `json:"occurred_at" binding:"required"`
}

type UpdateRecordRequest struct {
	Note string `json:"note"`
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSClient WebSocket 客户端
type WSClient struct {
	UserID int64
	Send   chan []byte
}
