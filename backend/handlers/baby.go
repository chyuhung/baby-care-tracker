package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"net/http"
	"strconv"

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

	// 今日喂奶次数
	var feedingCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM feeding_records WHERE baby_id = ? AND date(occurred_at, 'localtime') = date('now', 'localtime')",
		babyID,
	).Scan(&feedingCount)

	// 今日尿布次数
	var diaperCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM diaper_records WHERE baby_id = ? AND date(occurred_at, 'localtime') = date('now', 'localtime')",
		babyID,
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
		"SELECT COALESCE(SUM(amount_ml), 0) FROM feeding_records WHERE baby_id = ? AND date(occurred_at, 'localtime') = date('now', 'localtime') AND amount_ml > 0",
		babyID,
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

	// 查询最近7天的数据
	rows, err := database.DB.Query(`
		WITH dates AS (
			SELECT date('now', '-6 days', 'localtime') as date
			UNION ALL SELECT date('now', '-5 days', 'localtime')
			UNION ALL SELECT date('now', '-4 days', 'localtime')
			UNION ALL SELECT date('now', '-3 days', 'localtime')
			UNION ALL SELECT date('now', '-2 days', 'localtime')
			UNION ALL SELECT date('now', '-1 days', 'localtime')
			UNION ALL SELECT date('now', 'localtime')
		)
		SELECT 
			d.date,
			COALESCE(f.cnt, 0) as feeding_count,
			COALESCE(dp.cnt, 0) as diaper_count,
			COALESCE(f.ml, 0) as total_ml
		FROM dates d
		LEFT JOIN (
			SELECT date(occurred_at, 'localtime') as d, COUNT(*) as cnt, SUM(amount_ml) as ml
			FROM feeding_records 
			WHERE baby_id = ?
			GROUP BY date(occurred_at, 'localtime')
		) f ON d.date = f.d
		LEFT JOIN (
			SELECT date(occurred_at, 'localtime') as d, COUNT(*) as cnt
			FROM diaper_records 
			WHERE baby_id = ?
			GROUP BY date(occurred_at, 'localtime')
		) dp ON d.date = dp.d
		ORDER BY d.date ASC
	`, babyID, babyID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var trends []DailyStats
	for rows.Next() {
		var ds DailyStats
		if err := rows.Scan(&ds.Date, &ds.FeedingCount, &ds.DiaperCount, &ds.TotalMl); err != nil {
			continue
		}
		trends = append(trends, ds)
	}

	c.JSON(http.StatusOK, trends)
}
