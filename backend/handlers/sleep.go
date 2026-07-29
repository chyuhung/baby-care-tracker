package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func StartSleep(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.CreateSleepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.StartedAt == "" {
		req.StartedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	result, err := database.DB.Exec(
		"INSERT INTO sleep_records (baby_id, user_id, started_at, note) VALUES (?, ?, ?, ?)",
		babyID, userID, req.StartedAt, req.Note,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建睡眠记录失败"})
		return
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建睡眠记录失败"})
		return
	}

	var record models.SleepRecord
	err = database.DB.QueryRow(
		"SELECT id, baby_id, user_id, started_at, ended_at, note, created_at FROM sleep_records WHERE id = ?",
		recordID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.StartedAt, &record.EndedAt, &record.Note, &record.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建睡眠记录失败"})
		return
	}
	record.RecordType = "sleep"

	rec := models.Record{
		ID:         record.ID,
		BabyID:     record.BabyID,
		UserID:     record.UserID,
		RecordType: "sleep",
		Data:       record,
		OccurredAt: record.StartedAt,
		CreatedAt:  record.CreatedAt,
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "record_created",
		Payload: rec,
	})

	c.JSON(http.StatusCreated, rec)
}

func StopSleep(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	sleepID, err := parseInt64(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的睡眠ID"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.StopSleepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.EndedAt == "" {
		req.EndedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	_, err = database.DB.Exec(
		"UPDATE sleep_records SET ended_at = ?, note = ? WHERE id = ? AND ended_at IS NULL",
		req.EndedAt, req.Note, sleepID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "结束睡眠失败"})
		return
	}

	var record models.SleepRecord
	database.DB.QueryRow(
		"SELECT id, baby_id, user_id, started_at, ended_at, note, created_at FROM sleep_records WHERE id = ?",
		sleepID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.StartedAt, &record.EndedAt, &record.Note, &record.CreatedAt)
	record.RecordType = "sleep"

	rec := models.Record{
		ID:         record.ID,
		BabyID:     record.BabyID,
		UserID:     record.UserID,
		RecordType: "sleep",
		Data:       record,
		OccurredAt: record.StartedAt,
		CreatedAt:  record.CreatedAt,
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "record_created",
		Payload: rec,
	})

	c.JSON(http.StatusOK, rec)
}

func GetCurrentSleep(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var record models.SleepRecord
	err := database.DB.QueryRow(
		"SELECT id, baby_id, user_id, started_at, ended_at, note, created_at FROM sleep_records WHERE baby_id = ? AND ended_at IS NULL ORDER BY started_at DESC LIMIT 1",
		babyID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.StartedAt, &record.EndedAt, &record.Note, &record.CreatedAt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	record.RecordType = "sleep"

	c.JSON(http.StatusOK, record)
}
