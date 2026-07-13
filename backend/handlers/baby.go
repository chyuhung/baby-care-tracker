package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// checkBabyFamily 检查宝宝是否属于当前用户的家庭
func checkBabyFamily(babyID, userID int64) bool {
	var familyID int64
	database.DB.QueryRow(
		"SELECT u.family_id FROM babies b JOIN users u ON b.user_id = u.id WHERE b.id = ?",
		babyID,
	).Scan(&familyID)
	if familyID == 0 {
		return false
	}
	var userFamilyID int64
	database.DB.QueryRow("SELECT family_id FROM users WHERE id = ?", userID).Scan(&userFamilyID)
	return familyID == userFamilyID
}

// GetBabies 获取当前家庭的所有宝宝
func GetBabies(c *gin.Context) {
	userID := c.GetInt64("user_id")

	rows, err := database.DB.Query(
		`SELECT b.id, b.user_id, b.name, b.birth_date, b.gender, b.avatar_color, b.created_at
		FROM babies b
		JOIN users u ON b.user_id = u.id
		WHERE u.family_id = (SELECT family_id FROM users WHERE id = ?)
		ORDER BY b.created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var babies []models.Baby
	for rows.Next() {
		var b models.Baby
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.BirthDate, &b.Gender, &b.AvatarColor, &b.CreatedAt); err != nil {
			continue
		}
		babies = append(babies, b)
	}

	if babies == nil {
		babies = []models.Baby{}
	}

	c.JSON(http.StatusOK, babies)
}

// CreateBaby 创建宝宝
func CreateBaby(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.CreateBabyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "宝宝姓名和出生日期必填"})
		return
	}

	if req.AvatarColor == "" {
		req.AvatarColor = "#6C63FF"
	}

	result, err := database.DB.Exec(
		"INSERT INTO babies (user_id, name, birth_date, gender, avatar_color) VALUES (?, ?, ?, ?, ?)",
		userID, req.Name, req.BirthDate, req.Gender, req.AvatarColor,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建宝宝失败"})
		return
	}

	babyID, _ := result.LastInsertId()

	var baby models.Baby
	database.DB.QueryRow(
		"SELECT id, user_id, name, birth_date, gender, avatar_color, created_at FROM babies WHERE id = ?",
		babyID,
	).Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.BirthDate, &baby.Gender, &baby.AvatarColor, &baby.CreatedAt)

	// 广播给所有连接的客户端
	BroadcastMessage(models.WebSocketMessage{
		Type:    "baby_created",
		Payload: baby,
	})

	c.JSON(http.StatusCreated, baby)
}

// UpdateBaby 更新宝宝
func UpdateBaby(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req models.UpdateBabyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
		return
	}

	_, err := database.DB.Exec(
		"UPDATE babies SET name = COALESCE(NULLIF(?, ''), name), birth_date = COALESCE(NULLIF(?, ''), birth_date), gender = COALESCE(NULLIF(?, ''), gender), avatar_color = COALESCE(NULLIF(?, ''), avatar_color) WHERE id = ?",
		req.Name, req.BirthDate, req.Gender, req.AvatarColor, babyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	var baby models.Baby
	database.DB.QueryRow(
		"SELECT id, user_id, name, birth_date, gender, avatar_color, created_at FROM babies WHERE id = ?",
		babyID,
	).Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.BirthDate, &baby.Gender, &baby.AvatarColor, &baby.CreatedAt)

	BroadcastMessage(models.WebSocketMessage{
		Type:    "baby_updated",
		Payload: baby,
	})

	c.JSON(http.StatusOK, baby)
}

// DeleteBaby 删除宝宝
func DeleteBaby(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
		return
	}

	_, err := database.DB.Exec("DELETE FROM babies WHERE id = ?", babyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "baby_deleted",
		Payload: map[string]int64{"id": babyID},
	})

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetStats 获取宝宝统计
func GetStats(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	tzOffset := getTzOffset(c)
	todayStart, todayEnd := todayDateRange(tzOffset)

	// 今日喂奶次数
	var feedingCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM feeding_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ?",
		babyID, todayStart, todayEnd,
	).Scan(&feedingCount)

	// 今日尿布次数
	var diaperCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM diaper_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ?",
		babyID, todayStart, todayEnd,
	).Scan(&diaperCount)

	// 最后一次喂奶
	var lastFeeding string
	database.DB.QueryRow(
		"SELECT occurred_at FROM feeding_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&lastFeeding)

	// 最后一次尿布
	var lastDiaper string
	database.DB.QueryRow(
		"SELECT occurred_at FROM diaper_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&lastDiaper)

	// 今日总喂奶量
	var totalMl int
	database.DB.QueryRow(
		"SELECT COALESCE(SUM(amount_ml), 0) FROM feeding_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ? AND amount_ml > 0",
		babyID, todayStart, todayEnd,
	).Scan(&totalMl)

	c.JSON(http.StatusOK, gin.H{
		"feeding_count": feedingCount,
		"diaper_count":  diaperCount,
		"last_feeding":  lastFeeding,
		"last_diaper":   lastDiaper,
		"total_ml_today": totalMl,
	})
}

// DailyStats 每日统计数据结构
type DailyStats struct {
	Date        string `json:"date"`
	FeedingCount int    `json:"feeding_count"`
	DiaperCount  int    `json:"diaper_count"`
	TotalMl      int    `json:"total_ml"`
}

// GetTrendStats 获取宝宝趋势统计（最近7天）
func GetTrendStats(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
		return
	}

	tzOffset := getTzOffset(c)
	dates := lastNDates(tzOffset, 7)
	if len(dates) == 0 {
		c.JSON(http.StatusOK, []DailyStats{})
		return
	}

	startDate := daysAgoUTC(tzOffset, 7)

	// 在 Go 中按客户端时区分组，避免 SQLite date() 函数对时区修饰符支持不一致
	loc := time.FixedZone("user", tzOffset*60)

	// 查询7天内的喂奶原始数据
	feedingRows, err := database.DB.Query(`
		SELECT occurred_at, amount_ml FROM feeding_records
		WHERE baby_id = ? AND occurred_at >= ?
	`, babyID, startDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer feedingRows.Close()

	feedingMap := make(map[string]*DailyStats)
	for feedingRows.Next() {
		var occurredAt string
		var ml int
		if feedingRows.Scan(&occurredAt, &ml) != nil {
			continue
		}
		t := parseTime(occurredAt).In(loc)
		date := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
		ds, ok := feedingMap[date]
		if !ok {
			ds = &DailyStats{}
			feedingMap[date] = ds
		}
		ds.FeedingCount++
		ds.TotalMl += ml
	}

	// 查询7天内的尿布原始数据
	diaperRows, err := database.DB.Query(`
		SELECT occurred_at FROM diaper_records
		WHERE baby_id = ? AND occurred_at >= ?
	`, babyID, startDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer diaperRows.Close()

	diaperMap := make(map[string]int)
	for diaperRows.Next() {
		var occurredAt string
		if diaperRows.Scan(&occurredAt) != nil {
			continue
		}
		t := parseTime(occurredAt).In(loc)
		date := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
		diaperMap[date]++
	}

	var trends []DailyStats
	for _, date := range dates {
		f := feedingMap[date]
		trends = append(trends, DailyStats{
			Date:         date,
			FeedingCount: func() int { if f != nil { return f.FeedingCount }; return 0 }(),
			DiaperCount:  diaperMap[date],
			TotalMl:      func() int { if f != nil { return f.TotalMl }; return 0 }(),
		})
	}

	c.JSON(http.StatusOK, trends)
}
