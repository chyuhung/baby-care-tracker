package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func StartOutdoor(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.CreateOutdoorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.StartedAt == "" {
		req.StartedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	result, err := database.DB.Exec(
		"INSERT INTO outdoor_records (baby_id, user_id, started_at, note) VALUES (?, ?, ?, ?)",
		babyID, userID, req.StartedAt, req.Note,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var record models.OutdoorRecord
	var endedAt sql.NullString
	err = database.DB.QueryRow(
		"SELECT id, baby_id, user_id, started_at, ended_at, note, created_at FROM outdoor_records WHERE id = ?",
		recordID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.StartedAt, &endedAt, &record.Note, &record.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if endedAt.Valid {
		record.EndedAt = &endedAt.String
	}
	record.RecordType = "outdoor"

	rec := models.Record{
		ID:         record.ID,
		BabyID:     record.BabyID,
		UserID:     record.UserID,
		RecordType: "outdoor",
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

func StopOutdoor(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	outdoorID, err := parseInt64(c.Param("oid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的户外活动ID"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.StopOutdoorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.EndedAt == "" {
		req.EndedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	_, err = database.DB.Exec(
		"UPDATE outdoor_records SET ended_at = ?, note = ? WHERE id = ? AND ended_at IS NULL",
		req.EndedAt, req.Note, outdoorID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var record models.OutdoorRecord
	var endedAt2 sql.NullString
	database.DB.QueryRow(
		"SELECT id, baby_id, user_id, started_at, ended_at, note, created_at FROM outdoor_records WHERE id = ?",
		outdoorID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.StartedAt, &endedAt2, &record.Note, &record.CreatedAt)
	if endedAt2.Valid {
		record.EndedAt = &endedAt2.String
	}
	record.RecordType = "outdoor"

	rec := models.Record{
		ID:         record.ID,
		BabyID:     record.BabyID,
		UserID:     record.UserID,
		RecordType: "outdoor",
		Data:       record,
		OccurredAt: *record.EndedAt,
		CreatedAt:  record.CreatedAt,
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "record_created",
		Payload: rec,
	})

	c.JSON(http.StatusOK, rec)
}

func GetCurrentOutdoor(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var record models.OutdoorRecord
	var endedAt3 sql.NullString
	err := database.DB.QueryRow(
		"SELECT id, baby_id, user_id, started_at, ended_at, note, created_at FROM outdoor_records WHERE baby_id = ? AND ended_at IS NULL ORDER BY started_at DESC LIMIT 1",
		babyID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.StartedAt, &endedAt3, &record.Note, &record.CreatedAt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if endedAt3.Valid {
		record.EndedAt = &endedAt3.String
	}
	record.RecordType = "outdoor"

	c.JSON(http.StatusOK, record)
}
