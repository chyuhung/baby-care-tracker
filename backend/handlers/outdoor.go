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
	}, babyFamilyID(babyID))

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

	// 时序校验：先取 started_at，结束早于开始直接拒绝
	var outdoorStartedAt string
	qErr := database.DB.QueryRow(
		"SELECT started_at FROM outdoor_records WHERE id = ? AND baby_id = ?",
		outdoorID, babyID,
	).Scan(&outdoorStartedAt)
	if qErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在或已结束"})
		return
	}
	if parseTime(req.EndedAt).Before(parseTime(outdoorStartedAt)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
		return
	}

	res, err := database.DB.Exec(
		"UPDATE outdoor_records SET ended_at = ?, note = COALESCE(NULLIF(?, ''), note) WHERE id = ? AND baby_id = ? AND ended_at IS NULL",
		req.EndedAt, req.Note, outdoorID, babyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	// 谓词含 baby_id：别人宝宝的记录、或已结束的记录，一律视为不存在
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在或已结束"})
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
		OccurredAt: record.StartedAt,
		CreatedAt:  record.CreatedAt,
	}

	// 结束是「更新」不是「新建」：家人端据此原位替换该行并刷新统计
	BroadcastMessage(models.WebSocketMessage{
		Type:    "record_updated",
		Payload: rec,
	}, babyFamilyID(babyID))

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
