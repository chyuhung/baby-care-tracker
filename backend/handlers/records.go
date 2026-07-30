package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"database/sql"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

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
	switch recordType {
	case "diaper":
		database.DB.QueryRow("SELECT baby_id FROM diaper_records WHERE id = ?", recordID).Scan(&babyID)
	case "sleep":
		database.DB.QueryRow("SELECT baby_id FROM sleep_records WHERE id = ?", recordID).Scan(&babyID)
	case "temperature":
		database.DB.QueryRow("SELECT baby_id FROM temperature_records WHERE id = ?", recordID).Scan(&babyID)
	default:
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
	if recordType != "" && recordType != "feeding" && recordType != "diaper" && recordType != "sleep" && recordType != "temperature" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须为 feeding, diaper, sleep 或 temperature"})
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
	sleepDaysFilter := ""
	if daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 && days <= 365 {
			start := daysAgoUTC(tzOffset, days)
			daysFilter = " AND occurred_at >= ?"
			sleepDaysFilter = " AND started_at >= ?"
			args = append(args, start)
		}
	}

	var feedingCount, diaperCount, sleepCount, temperatureCount int
	if recordType == "" || recordType == "feeding" {
		fArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM feeding_records WHERE baby_id = ?"+daysFilter, fArgs...).Scan(&feedingCount)
	}
	if recordType == "" || recordType == "diaper" {
		dArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM diaper_records WHERE baby_id = ?"+daysFilter, dArgs...).Scan(&diaperCount)
	}
	if recordType == "" || recordType == "sleep" {
		sArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM sleep_records WHERE baby_id = ? AND ended_at IS NOT NULL"+sleepDaysFilter, sArgs...).Scan(&sleepCount)
	}
	if recordType == "" || recordType == "temperature" {
		tArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM temperature_records WHERE baby_id = ?"+daysFilter, tArgs...).Scan(&temperatureCount)
	}
	c.Header("X-Total-Count", strconv.Itoa(feedingCount+diaperCount+sleepCount+temperatureCount))

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

	if recordType == "" || recordType == "sleep" {
		sArgs := append([]interface{}{}, args...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, started_at, ended_at, note, created_at
			FROM sleep_records WHERE baby_id = ? AND ended_at IS NOT NULL`+sleepDaysFilter+` ORDER BY ended_at DESC LIMIT 500`,
			sArgs...,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.SleepRecord
				var note string
				var endedAt sql.NullString
				if err := rows.Scan(&r.ID, &r.BabyID, &r.UserID, &r.StartedAt, &endedAt, &note, &r.CreatedAt); err != nil {
					continue
				}
				if endedAt.Valid {
					r.EndedAt = &endedAt.String
				}
				r.Note = note
				r.RecordType = "sleep"
				records = append(records, models.Record{
					ID:         r.ID,
					BabyID:     r.BabyID,
					UserID:     r.UserID,
					RecordType: "sleep",
					Data:       r,
					OccurredAt: r.StartedAt,
					CreatedAt:  r.CreatedAt,
				})
			}
		}
	}

	if recordType == "" || recordType == "temperature" {
		tArgs := append([]interface{}{}, args...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, temperature, location, note, occurred_at, created_at
			FROM temperature_records WHERE baby_id = ?`+daysFilter+` ORDER BY occurred_at DESC LIMIT 500`,
			tArgs...,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.TemperatureRecord
				var note, location string
				var temp float64
				if err := rows.Scan(&r.ID, &r.BabyID, &r.UserID, &temp, &location, &note, &r.OccurredAt, &r.CreatedAt); err != nil {
					continue
				}
				r.Temperature = temp
				r.Location = location
				r.Note = note
				r.RecordType = "temperature"
				records = append(records, models.Record{
					ID:         r.ID,
					BabyID:     r.BabyID,
					UserID:     r.UserID,
					RecordType: "temperature",
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

	var feedingCount, diaperCount, sleepCount, temperatureCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM feeding_records WHERE baby_id = ?", babyID).Scan(&feedingCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM diaper_records WHERE baby_id = ?", babyID).Scan(&diaperCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM sleep_records WHERE baby_id = ? AND ended_at IS NOT NULL", babyID).Scan(&sleepCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM temperature_records WHERE baby_id = ?", babyID).Scan(&temperatureCount)

	c.JSON(http.StatusOK, gin.H{
		"feeding_count":     feedingCount,
		"diaper_count":      diaperCount,
		"sleep_count":       sleepCount,
		"temperature_count": temperatureCount,
		"total":             feedingCount + diaperCount + sleepCount + temperatureCount,
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

	switch recordType {
	case "diaper":
		_, err := database.DB.Exec(
			"UPDATE diaper_records SET type = ?, note = ?, occurred_at = ? WHERE id = ?",
			req.Type, req.Note, req.OccurredAt, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	case "sleep":
		_, err := database.DB.Exec(
			"UPDATE sleep_records SET started_at = ?, ended_at = ?, note = ? WHERE id = ?",
			req.StartedAt, req.EndedAt, req.Note, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	case "temperature":
		_, err := database.DB.Exec(
			"UPDATE temperature_records SET temperature = ?, location = ?, note = ?, occurred_at = ? WHERE id = ?",
			req.Temperature, req.Location, req.Note, req.OccurredAt, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	default:
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

	var err error
	switch recordType {
	case "diaper":
		_, err = database.DB.Exec("DELETE FROM diaper_records WHERE id = ?", recordID)
	case "sleep":
		_, err = database.DB.Exec("DELETE FROM sleep_records WHERE id = ?", recordID)
	case "temperature":
		_, err = database.DB.Exec("DELETE FROM temperature_records WHERE id = ?", recordID)
	default:
		_, err = database.DB.Exec("DELETE FROM feeding_records WHERE id = ?", recordID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
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
