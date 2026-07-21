package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return 0, false
	}
	return id, true
}

func lookupBabyID(recordID int64, recordType string) int64 {
	var babyID int64
	if recordType == "diaper" {
		database.DB.QueryRow("SELECT baby_id FROM diaper_records WHERE id = ?", recordID).Scan(&babyID)
	} else {
		database.DB.QueryRow("SELECT baby_id FROM feeding_records WHERE id = ?", recordID).Scan(&babyID)
	}
	return babyID
}

// GetRecords 获取某宝宝所有记录（统一时间线）
func GetRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	recordType := c.Query("type")
	if recordType != "" && recordType != "feeding" && recordType != "diaper" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须为 feeding 或 diaper"})
		return
	}
	daysStr := c.Query("days")

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	tzOffset := getTzOffset(c)
	args := []interface{}{babyID}
	daysFilter := ""
	if daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 && days <= 365 {
			start := daysAgoUTC(tzOffset, days)
			daysFilter = " AND occurred_at >= ?"
			args = append(args, start)
		}
	}

	var feedingCount, diaperCount int
	if recordType == "" || recordType == "feeding" {
		fArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM feeding_records WHERE baby_id = ?"+daysFilter, fArgs...).Scan(&feedingCount)
	}
	if recordType == "" || recordType == "diaper" {
		dArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM diaper_records WHERE baby_id = ?"+daysFilter, dArgs...).Scan(&diaperCount)
	}
	c.Header("X-Total-Count", strconv.Itoa(feedingCount+diaperCount))

	var records []models.Record

	if recordType == "" || recordType == "feeding" {
		fArgs := append([]interface{}{}, args...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, type, duration_minutes, amount_ml, side, brand, note, occurred_at, created_at
			FROM feeding_records WHERE baby_id = ?`+daysFilter+` ORDER BY occurred_at DESC LIMIT 500`,
			fArgs...,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.FeedingRecord
				var note, brand, side string
				var duration, amount int
				if err := rows.Scan(&r.ID, &r.BabyID, &r.UserID, &r.Type, &duration, &amount, &side, &brand, &note, &r.OccurredAt, &r.CreatedAt); err != nil {
					continue
				}
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

	if recordType == "" || recordType == "diaper" {
		dArgs := append([]interface{}{}, args...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, type, note, occurred_at, created_at
			FROM diaper_records WHERE baby_id = ?`+daysFilter+` ORDER BY occurred_at DESC LIMIT 500`,
			dArgs...,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.DiaperRecord
				var note string
				if err := rows.Scan(&r.ID, &r.BabyID, &r.UserID, &r.Type, &note, &r.OccurredAt, &r.CreatedAt); err != nil {
					continue
				}
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
	} else {
		sort.Slice(records, func(i, j int) bool {
			ti := parseTime(records[i].OccurredAt)
			tj := parseTime(records[j].OccurredAt)
			return ti.After(tj)
		})
	}

	c.JSON(http.StatusOK, records)
}

// GetRecordsCount 获取宝宝记录总数
func GetRecordsCount(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var feedingCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM feeding_records WHERE baby_id = ?", babyID).Scan(&feedingCount)
	var diaperCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM diaper_records WHERE baby_id = ?", babyID).Scan(&diaperCount)

	c.JSON(http.StatusOK, gin.H{
		"feeding_count": feedingCount,
		"diaper_count":  diaperCount,
		"total":         feedingCount + diaperCount,
	})
}

// CreateFeeding 创建喂奶记录
func CreateFeeding(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.CreateFeedingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.OccurredAt == "" {
		req.OccurredAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
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

	recordID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败"})
		return
	}

	var record models.FeedingRecord
	err = database.DB.QueryRow(
		"SELECT id, baby_id, user_id, type, duration_minutes, amount_ml, side, brand, note, occurred_at, created_at FROM feeding_records WHERE id = ?",
		recordID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.Type, &record.DurationMinutes, &record.AmountMl, &record.Side, &record.Brand, &record.Note, &record.OccurredAt, &record.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败"})
		return
	}
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
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.CreateDiaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.OccurredAt == "" {
		req.OccurredAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	result, err := database.DB.Exec(
		"INSERT INTO diaper_records (baby_id, user_id, type, note, occurred_at) VALUES (?, ?, ?, ?, ?)",
		babyID, userID, req.Type, req.Note, req.OccurredAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败"})
		return
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败"})
		return
	}

	var record models.DiaperRecord
	err = database.DB.QueryRow(
		"SELECT id, baby_id, user_id, type, note, occurred_at, created_at FROM diaper_records WHERE id = ?",
		recordID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.Type, &record.Note, &record.OccurredAt, &record.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败"})
		return
	}
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
	recordID, ok := parseID(c)
	if !ok {
		return
	}
	recordType := c.Query("type")

	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	babyID := lookupBabyID(recordID, recordType)
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	if recordType == "diaper" {
		_, err := database.DB.Exec(
			"UPDATE diaper_records SET type = ?, note = ?, occurred_at = ? WHERE id = ?",
			req.Type, req.Note, req.OccurredAt, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	} else {
		_, err := database.DB.Exec(
			"UPDATE feeding_records SET type = ?, duration_minutes = ?, amount_ml = ?, side = ?, brand = ?, note = ?, occurred_at = ? WHERE id = ?",
			req.Type, req.DurationMinutes, req.AmountMl, req.Side, req.Brand, req.Note, req.OccurredAt, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteRecord 删除记录
func DeleteRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")
	recordID, ok := parseID(c)
	if !ok {
		return
	}
	recordType := c.Query("type")

	babyID := lookupBabyID(recordID, recordType)
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	if recordType == "diaper" {
		_, err := database.DB.Exec("DELETE FROM diaper_records WHERE id = ?", recordID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
			return
		}
	} else {
		_, err := database.DB.Exec("DELETE FROM feeding_records WHERE id = ?", recordID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
			return
		}
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
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var record models.FeedingRecord
	var note, brand, side string
	var duration, amount int
	err := database.DB.QueryRow(
		"SELECT type, duration_minutes, amount_ml, side, brand, note FROM feeding_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&record.Type, &duration, &amount, &side, &brand, &note)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":             record.Type,
		"duration_minutes": duration,
		"amount_ml":        amount,
		"side":             side,
		"brand":            brand,
		"note":             note,
	})
}
