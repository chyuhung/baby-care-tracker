package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetRecords 获取某宝宝所有记录（统一时间线）
func GetRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	recordType := c.Query("type") // optional filter

	// 验证宝宝归属
	var ownerID int64
	database.DB.QueryRow("SELECT user_id FROM babies WHERE id = ?", babyID).Scan(&ownerID)
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var records []models.Record

	// 喂奶记录
	if recordType == "" || recordType == "feeding" {
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, type, duration_minutes, amount_ml, side, brand, note, occurred_at, created_at
			FROM feeding_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 200`,
			babyID,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.FeedingRecord
				var note, brand, side string
				var duration, amount int
				rows.Scan(&r.ID, &r.BabyID, &r.UserID, &r.Type, &duration, &amount, &side, &brand, &note, &r.OccurredAt, &r.CreatedAt)
				r.Note = note
				r.Brand = brand
				r.Side = side
				r.DurationMinutes = duration
				r.AmountMl = amount
				r.RecordType = "feeding"
				records = append(records, models.Record{
					ID:         r.ID,
					BabyID:     r.BabyID,
					UserID:     r.UserID,
					RecordType: "feeding",
					Data:       r,
					OccurredAt: r.OccurredAt,
					CreatedAt:  r.CreatedAt,
				})
			}
		}
	}

	// 尿布记录
	if recordType == "" || recordType == "diaper" {
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, type, note, occurred_at, created_at
			FROM diaper_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 200`,
			babyID,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.DiaperRecord
				var note string
				rows.Scan(&r.ID, &r.BabyID, &r.UserID, &r.Type, &note, &r.OccurredAt, &r.CreatedAt)
				r.Note = note
				r.RecordType = "diaper"
				records = append(records, models.Record{
					ID:         r.ID,
					BabyID:     r.BabyID,
					UserID:     r.UserID,
					RecordType: "diaper",
					Data:       r,
					OccurredAt: r.OccurredAt,
					CreatedAt:  r.CreatedAt,
				})
			}
		}
	}

	if records == nil {
		records = []models.Record{}
	}

	c.JSON(http.StatusOK, records)
}

// CreateFeeding 创建喂奶记录
func CreateFeeding(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var ownerID int64
	database.DB.QueryRow("SELECT user_id FROM babies WHERE id = ?", babyID).Scan(&ownerID)
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.CreateFeedingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.OccurredAt == "" {
		req.OccurredAt = time.Now().Format("2006-01-02T15:04:05")
	}

	result, err := database.DB.Exec(
		`INSERT INTO feeding_records (baby_id, user_id, type, duration_minutes, amount_ml, side, brand, note, occurred_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		babyID, userID, req.Type, req.DurationMinutes, req.AmountMl, req.Side, req.Brand, req.Note, req.OccurredAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败"})
		return
	}

	recordID, _ := result.LastInsertId()
	var record models.FeedingRecord
	database.DB.QueryRow(
		"SELECT id, baby_id, user_id, type, duration_minutes, amount_ml, side, brand, note, occurred_at, created_at FROM feeding_records WHERE id = ?",
		recordID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.Type, &record.DurationMinutes, &record.AmountMl, &record.Side, &record.Brand, &record.Note, &record.OccurredAt, &record.CreatedAt)
	record.RecordType = "feeding"

	rec := models.Record{
		ID:         record.ID,
		BabyID:     record.BabyID,
		UserID:     record.UserID,
		RecordType: "feeding",
		Data:       record,
		OccurredAt: record.OccurredAt,
		CreatedAt:  record.CreatedAt,
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "record_created",
		Payload: rec,
	})

	c.JSON(http.StatusCreated, rec)
}

// CreateDiaper 创建尿布记录
func CreateDiaper(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var ownerID int64
	database.DB.QueryRow("SELECT user_id FROM babies WHERE id = ?", babyID).Scan(&ownerID)
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.CreateDiaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.OccurredAt == "" {
		req.OccurredAt = time.Now().Format("2006-01-02T15:04:05")
	}

	result, err := database.DB.Exec(
		"INSERT INTO diaper_records (baby_id, user_id, type, note, occurred_at) VALUES (?, ?, ?, ?, ?)",
		babyID, userID, req.Type, req.Note, req.OccurredAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败"})
		return
	}

	recordID, _ := result.LastInsertId()
	var record models.DiaperRecord
	database.DB.QueryRow(
		"SELECT id, baby_id, user_id, type, note, occurred_at, created_at FROM diaper_records WHERE id = ?",
		recordID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.Type, &record.Note, &record.OccurredAt, &record.CreatedAt)
	record.RecordType = "diaper"

	rec := models.Record{
		ID:         record.ID,
		BabyID:     record.BabyID,
		UserID:     record.UserID,
		RecordType: "diaper",
		Data:       record,
		OccurredAt: record.OccurredAt,
		CreatedAt:  record.CreatedAt,
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "record_created",
		Payload: rec,
	})

	c.JSON(http.StatusCreated, rec)
}

// UpdateRecord 更新记录
func UpdateRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")
	recordID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	recordType := c.Query("type")

	var req models.UpdateRecordRequest
	c.ShouldBindJSON(&req)

	// 验证归属
	var ownerID int64
	if recordType == "diaper" {
		database.DB.QueryRow("SELECT user_id FROM diaper_records WHERE id = ?", recordID).Scan(&ownerID)
	} else {
		database.DB.QueryRow("SELECT user_id FROM feeding_records WHERE id = ?", recordID).Scan(&ownerID)
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	if recordType == "diaper" {
		database.DB.Exec("UPDATE diaper_records SET note = ? WHERE id = ?", req.Note, recordID)
	} else {
		database.DB.Exec("UPDATE feeding_records SET note = ? WHERE id = ?", req.Note, recordID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteRecord 删除记录
func DeleteRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")
	recordID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	recordType := c.Query("type")

	// 验证归属
	var ownerID int64
	if recordType == "diaper" {
		database.DB.QueryRow("SELECT user_id FROM diaper_records WHERE id = ?", recordID).Scan(&ownerID)
	} else {
		database.DB.QueryRow("SELECT user_id FROM feeding_records WHERE id = ?", recordID).Scan(&ownerID)
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	if recordType == "diaper" {
		database.DB.Exec("DELETE FROM diaper_records WHERE id = ?", recordID)
	} else {
		database.DB.Exec("DELETE FROM feeding_records WHERE id = ?", recordID)
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "record_deleted",
		Payload: map[string]interface{}{"id": recordID, "type": recordType},
	})

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetLatestFeeding 获取最近一次喂奶记录（用于快捷填表）
func GetLatestFeeding(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var record models.FeedingRecord
	var note, brand, side string
	var duration, amount int
	err := database.DB.QueryRow(
		"SELECT type, duration_minutes, amount_ml, side, brand, note FROM feeding_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&record.Type, &duration, &amount, &side, &brand, &note)

	if err != nil {
		_ = userID
		c.JSON(http.StatusOK, nil)
		return
	}

	_ = userID
	c.JSON(http.StatusOK, gin.H{
		"type":             record.Type,
		"duration_minutes": duration,
		"amount_ml":        amount,
		"side":             side,
		"brand":            brand,
		"note":             note,
	})
}
