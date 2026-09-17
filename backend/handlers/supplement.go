package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateSupplement 创建补剂记录
func CreateSupplement(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var req models.CreateSupplementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.OccurredAt == "" {
		req.OccurredAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	result, err := database.DB.Exec(
		"INSERT INTO supplement_records (baby_id, user_id, name, dosage_value, dosage_unit, note, occurred_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		babyID, userID, req.Name, req.DosageValue, req.DosageUnit, req.Note, req.OccurredAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建补剂记录失败"})
		return
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建补剂记录失败"})
		return
	}

	var record models.SupplementRecord
	err = database.DB.QueryRow(
		"SELECT id, baby_id, user_id, name, dosage_value, dosage_unit, note, occurred_at, created_at FROM supplement_records WHERE id = ?",
		recordID,
	).Scan(&record.ID, &record.BabyID, &record.UserID, &record.Name, &record.DosageValue, &record.DosageUnit, &record.Note, &record.OccurredAt, &record.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建补剂记录失败"})
		return
	}
	record.RecordType = "supplement"

	rec := models.Record{
		ID:         record.ID,
		BabyID:     record.BabyID,
		UserID:     record.UserID,
		RecordType: "supplement",
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

// GetLatestSupplement 获取最近一次补剂记录（用于快捷填表）
func GetLatestSupplement(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var name, dosageUnit, note string
	var dosageValue float64
	err := database.DB.QueryRow(
		"SELECT name, dosage_value, dosage_unit, note FROM supplement_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&name, &dosageValue, &dosageUnit, &note)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":         name,
		"dosage_value": dosageValue,
		"dosage_unit":  dosageUnit,
		"note":         note,
	})
}