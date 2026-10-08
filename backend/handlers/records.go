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
	case "outdoor":
		database.DB.QueryRow("SELECT baby_id FROM outdoor_records WHERE id = ?", recordID).Scan(&babyID)
	case "supplement":
		database.DB.QueryRow("SELECT baby_id FROM supplement_records WHERE id = ?", recordID).Scan(&babyID)
	default:
		database.DB.QueryRow("SELECT baby_id FROM feeding_records WHERE id = ?", recordID).Scan(&babyID)
	}
	return babyID
}

// recordTables 记录类型 → 表名白名单。表名只能来自本映射，绝不拼接用户输入。
var recordTables = map[string]string{
	"feeding":     "feeding_records",
	"diaper":      "diaper_records",
	"sleep":       "sleep_records",
	"temperature": "temperature_records",
	"outdoor":     "outdoor_records",
	"supplement":  "supplement_records",
}

// GetRecord 按 id + type 读取单条记录（编辑页加载用）。
// 此前编辑页没有单条读取端点，只能从列表窗口反查——记录不在窗口内时点「编辑」毫无反应。
func GetRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseInt64(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	recordType := c.Query("type")
	table, ok := recordTables[recordType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须为 feeding, diaper, sleep, temperature, outdoor 或 supplement"})
		return
	}
	var babyID int64
	if err := database.DB.QueryRow("SELECT baby_id FROM "+table+" WHERE id = ?", id).Scan(&babyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	rec, ok := loadSingleRecord(recordType, id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}
	c.JSON(http.StatusOK, rec)
}

// loadSingleRecord 按类型读取单条记录并组装为统一 Record（扫描逻辑与 GetRecords 列表一致）
func loadSingleRecord(recordType string, id int64) (models.Record, bool) {
	var rec models.Record
	var err error
	switch recordType {
	case "feeding":
		var r models.FeedingRecord
		err = database.DB.QueryRow(
			"SELECT id, baby_id, user_id, type, duration_minutes, amount_ml, side, brand, note, occurred_at, created_at FROM feeding_records WHERE id = ?", id,
		).Scan(&r.ID, &r.BabyID, &r.UserID, &r.Type, &r.DurationMinutes, &r.AmountMl, &r.Side, &r.Brand, &r.Note, &r.OccurredAt, &r.CreatedAt)
		if err == nil {
			r.RecordType = "feeding"
			rec = models.Record{ID: r.ID, BabyID: r.BabyID, UserID: r.UserID, RecordType: "feeding", Data: r, OccurredAt: r.OccurredAt, CreatedAt: r.CreatedAt}
		}
	case "diaper":
		var r models.DiaperRecord
		err = database.DB.QueryRow(
			"SELECT id, baby_id, user_id, type, note, occurred_at, created_at FROM diaper_records WHERE id = ?", id,
		).Scan(&r.ID, &r.BabyID, &r.UserID, &r.Type, &r.Note, &r.OccurredAt, &r.CreatedAt)
		if err == nil {
			r.RecordType = "diaper"
			rec = models.Record{ID: r.ID, BabyID: r.BabyID, UserID: r.UserID, RecordType: "diaper", Data: r, OccurredAt: r.OccurredAt, CreatedAt: r.CreatedAt}
		}
	case "sleep", "outdoor":
		table := recordTables[recordType]
		var startedAt, note, createdAt string
		var endedAt sql.NullString
		var rid, rBabyID, rUserID int64
		err = database.DB.QueryRow(
			"SELECT id, baby_id, user_id, started_at, ended_at, note, created_at FROM "+table+" WHERE id = ?", id,
		).Scan(&rid, &rBabyID, &rUserID, &startedAt, &endedAt, &note, &createdAt)
		if err == nil {
			if recordType == "sleep" {
				r := models.SleepRecord{ID: rid, BabyID: rBabyID, UserID: rUserID, StartedAt: startedAt, Note: note, CreatedAt: createdAt, RecordType: "sleep"}
				if endedAt.Valid {
					r.EndedAt = &endedAt.String
				}
				rec = models.Record{ID: rid, BabyID: rBabyID, UserID: rUserID, RecordType: "sleep", Data: r, OccurredAt: startedAt, CreatedAt: createdAt}
			} else {
				r := models.OutdoorRecord{ID: rid, BabyID: rBabyID, UserID: rUserID, StartedAt: startedAt, Note: note, CreatedAt: createdAt, RecordType: "outdoor"}
				if endedAt.Valid {
					r.EndedAt = &endedAt.String
				}
				rec = models.Record{ID: rid, BabyID: rBabyID, UserID: rUserID, RecordType: "outdoor", Data: r, OccurredAt: startedAt, CreatedAt: createdAt}
			}
		}
	case "temperature":
		var r models.TemperatureRecord
		var location string
		err = database.DB.QueryRow(
			"SELECT id, baby_id, user_id, temperature, location, note, occurred_at, created_at FROM temperature_records WHERE id = ?", id,
		).Scan(&r.ID, &r.BabyID, &r.UserID, &r.Temperature, &location, &r.Note, &r.OccurredAt, &r.CreatedAt)
		if err == nil {
			r.Location = location
			r.RecordType = "temperature"
			rec = models.Record{ID: r.ID, BabyID: r.BabyID, UserID: r.UserID, RecordType: "temperature", Data: r, OccurredAt: r.OccurredAt, CreatedAt: r.CreatedAt}
		}
	case "supplement":
		var r models.SupplementRecord
		err = database.DB.QueryRow(
			"SELECT id, baby_id, user_id, name, dosage_value, dosage_unit, note, occurred_at, created_at FROM supplement_records WHERE id = ?", id,
		).Scan(&r.ID, &r.BabyID, &r.UserID, &r.Name, &r.DosageValue, &r.DosageUnit, &r.Note, &r.OccurredAt, &r.CreatedAt)
		if err == nil {
			r.RecordType = "supplement"
			rec = models.Record{ID: r.ID, BabyID: r.BabyID, UserID: r.UserID, RecordType: "supplement", Data: r, OccurredAt: r.OccurredAt, CreatedAt: r.CreatedAt}
		}
	default:
		return models.Record{}, false
	}
	if err != nil {
		return models.Record{}, false
	}
	return rec, true
}

// GetRecords 获取某宝宝所有记录（统一时间线）
func GetRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	recordType := c.Query("type")
	if recordType != "" && recordType != "feeding" && recordType != "diaper" && recordType != "sleep" && recordType != "temperature" && recordType != "outdoor" && recordType != "supplement" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须为 feeding, diaper, sleep, temperature, outdoor 或 supplement"})
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
	// 列表按「事件起始时刻」归属日期，故用 started_at >= 窗口起；
	// 这与统计接口的「区间重叠」窗口有意不同：
	// 一条 8 天前开始、2 天前结束的睡眠（忘记结束）应计入统计里那两天的分钟数，
	// 但不能在「最近 7 天」列表里显示成一条日期在窗口外的记录。
	// 列表一律包含进行中的记录（ended_at 为 NULL），与首页「进行中」实时计时一致。
	sleepDaysFilter := ""
	outdoorDaysFilter := ""
	if daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 && days <= 365 {
			start := daysAgoUTC(tzOffset, days)
			daysFilter = " AND occurred_at >= ?"
			sleepDaysFilter = " AND started_at >= ?"
			outdoorDaysFilter = " AND started_at >= ?"
			args = append(args, start)
		}
	}

	// 分页：limit>0 时启用全局窗口分页 [offset, offset+limit)。
	// 每类表一次性预取 offset+limit 条，合并排序后按窗口截断，
	// 保证每页条数固定、跨页不重不漏（gap-free）。
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	pageSQL := " LIMIT 500"
	pageArgs := []interface{}{}
	pageOffset := 0
	pageLimit := 0
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		if l > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit 不能超过 200"})
			return
		}
		if oo, err := strconv.Atoi(offsetStr); err == nil && oo > 0 {
			pageOffset = oo
		}
		pageLimit = l
		pageSQL = " LIMIT ?"
		pageArgs = append(pageArgs, pageOffset+pageLimit)
	}

	var feedingCount, diaperCount, sleepCount, temperatureCount, outdoorCount, supplementCount int
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
		database.DB.QueryRow("SELECT COUNT(*) FROM sleep_records WHERE baby_id = ?"+sleepDaysFilter, sArgs...).Scan(&sleepCount)
	}
	if recordType == "" || recordType == "temperature" {
		tArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM temperature_records WHERE baby_id = ?"+daysFilter, tArgs...).Scan(&temperatureCount)
	}
	if recordType == "" || recordType == "outdoor" {
		oArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM outdoor_records WHERE baby_id = ?"+outdoorDaysFilter, oArgs...).Scan(&outdoorCount)
	}
	if recordType == "" || recordType == "supplement" {
		sArgs := append([]interface{}{}, args...)
		database.DB.QueryRow("SELECT COUNT(*) FROM supplement_records WHERE baby_id = ?"+daysFilter, sArgs...).Scan(&supplementCount)
	}
	c.Header("X-Total-Count", strconv.Itoa(feedingCount+diaperCount+sleepCount+temperatureCount+outdoorCount+supplementCount))

	var records []models.Record

	if recordType == "" || recordType == "feeding" {
		fArgs := append(append([]interface{}{}, args...), pageArgs...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, type, duration_minutes, amount_ml, side, brand, note, occurred_at, created_at
			FROM feeding_records WHERE baby_id = ?`+daysFilter+` ORDER BY occurred_at DESC`+pageSQL,
			fArgs...,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
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
		dArgs := append(append([]interface{}{}, args...), pageArgs...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, type, note, occurred_at, created_at
			FROM diaper_records WHERE baby_id = ?`+daysFilter+` ORDER BY occurred_at DESC`+pageSQL,
			dArgs...,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
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
		sArgs := append(append([]interface{}{}, args...), pageArgs...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, started_at, ended_at, note, created_at
			FROM sleep_records WHERE baby_id = ?`+sleepDaysFilter+` ORDER BY started_at DESC`+pageSQL,
			sArgs...,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
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
		tArgs := append(append([]interface{}{}, args...), pageArgs...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, temperature, location, note, occurred_at, created_at
			FROM temperature_records WHERE baby_id = ?`+daysFilter+` ORDER BY occurred_at DESC`+pageSQL,
			tArgs...,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
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

	if recordType == "" || recordType == "outdoor" {
		oArgs := append(append([]interface{}{}, args...), pageArgs...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, started_at, ended_at, note, created_at
			FROM outdoor_records WHERE baby_id = ?`+outdoorDaysFilter+` ORDER BY started_at DESC`+pageSQL,
			oArgs...,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.OutdoorRecord
				var note string
				var endedAt sql.NullString
				if err := rows.Scan(&r.ID, &r.BabyID, &r.UserID, &r.StartedAt, &endedAt, &note, &r.CreatedAt); err != nil {
					continue
				}
				if endedAt.Valid {
					r.EndedAt = &endedAt.String
				}
				r.Note = note
				r.RecordType = "outdoor"
				records = append(records, models.Record{
					ID:         r.ID,
					BabyID:     r.BabyID,
					UserID:     r.UserID,
					RecordType: "outdoor",
					Data:       r,
					OccurredAt: r.StartedAt,
					CreatedAt:  r.CreatedAt,
				})
			}
		}
	}

	if recordType == "" || recordType == "supplement" {
		sArgs := append(append([]interface{}{}, args...), pageArgs...)
		rows, err := database.DB.Query(
			`SELECT id, baby_id, user_id, name, dosage_value, dosage_unit, note, occurred_at, created_at
			FROM supplement_records WHERE baby_id = ?`+daysFilter+` ORDER BY occurred_at DESC`+pageSQL,
			sArgs...,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r models.SupplementRecord
				var note, name, dosageUnit string
				var dosageValue float64
				if err := rows.Scan(&r.ID, &r.BabyID, &r.UserID, &name, &dosageValue, &dosageUnit, &note, &r.OccurredAt, &r.CreatedAt); err != nil {
					continue
				}
				r.Name = name
				r.DosageValue = dosageValue
				r.DosageUnit = dosageUnit
				r.Note = note
				r.RecordType = "supplement"
				records = append(records, models.Record{
					ID:         r.ID,
					BabyID:     r.BabyID,
					UserID:     r.UserID,
					RecordType: "supplement",
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

	if pageLimit > 0 {
		start := pageOffset
		if start > len(records) {
			start = len(records)
		}
		end := pageOffset + pageLimit
		if end > len(records) {
			end = len(records)
		}
		records = records[start:end]
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

	recordType := c.Query("type")
	if recordType != "" && recordType != "feeding" && recordType != "diaper" && recordType != "sleep" && recordType != "temperature" && recordType != "outdoor" && recordType != "supplement" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须为 feeding, diaper, sleep, temperature, outdoor 或 supplement"})
		return
	}

	countAll := recordType == ""
	var feedingCount, diaperCount, sleepCount, temperatureCount, outdoorCount, supplementCount int
	if countAll || recordType == "feeding" {
		database.DB.QueryRow("SELECT COUNT(*) FROM feeding_records WHERE baby_id = ?", babyID).Scan(&feedingCount)
	}
	if countAll || recordType == "diaper" {
		database.DB.QueryRow("SELECT COUNT(*) FROM diaper_records WHERE baby_id = ?", babyID).Scan(&diaperCount)
	}
	if countAll || recordType == "sleep" {
		database.DB.QueryRow("SELECT COUNT(*) FROM sleep_records WHERE baby_id = ?", babyID).Scan(&sleepCount)
	}
	if countAll || recordType == "temperature" {
		database.DB.QueryRow("SELECT COUNT(*) FROM temperature_records WHERE baby_id = ?", babyID).Scan(&temperatureCount)
	}
	if countAll || recordType == "outdoor" {
		database.DB.QueryRow("SELECT COUNT(*) FROM outdoor_records WHERE baby_id = ?", babyID).Scan(&outdoorCount)
	}
	if countAll || recordType == "supplement" {
		database.DB.QueryRow("SELECT COUNT(*) FROM supplement_records WHERE baby_id = ?", babyID).Scan(&supplementCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"feeding_count":     feedingCount,
		"diaper_count":      diaperCount,
		"sleep_count":       sleepCount,
		"temperature_count": temperatureCount,
		"outdoor_count":     outdoorCount,
		"supplement_count":  supplementCount,
		"total":             feedingCount + diaperCount + sleepCount + temperatureCount + outdoorCount + supplementCount,
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
	}, babyFamilyID(babyID))

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
	}, babyFamilyID(babyID))

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
	if _, ok := recordTables[recordType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须为 feeding, diaper, sleep, temperature, outdoor 或 supplement"})
		return
	}

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
		if req.EndedAt != "" && req.StartedAt != "" && parseTime(req.EndedAt).Before(parseTime(req.StartedAt)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
			return
		}
		_, err := database.DB.Exec(
			"UPDATE sleep_records SET started_at = ?, ended_at = CASE WHEN ? = '' THEN NULL ELSE ? END, note = ? WHERE id = ?",
			req.StartedAt, req.EndedAt, req.EndedAt, req.Note, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	case "temperature":
		if req.Temperature < 30 || req.Temperature > 60 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "体温需在 30-60°C 之间"})
			return
		}
		_, err := database.DB.Exec(
			"UPDATE temperature_records SET temperature = ?, location = ?, note = ?, occurred_at = ? WHERE id = ?",
			req.Temperature, req.Location, req.Note, req.OccurredAt, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	case "outdoor":
		if req.EndedAt != "" && req.StartedAt != "" && parseTime(req.EndedAt).Before(parseTime(req.StartedAt)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
			return
		}
		_, err := database.DB.Exec(
			"UPDATE outdoor_records SET started_at = ?, ended_at = CASE WHEN ? = '' THEN NULL ELSE ? END, note = ? WHERE id = ?",
			req.StartedAt, req.EndedAt, req.EndedAt, req.Note, recordID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	case "supplement":
		_, err := database.DB.Exec(
			"UPDATE supplement_records SET name = ?, dosage_value = ?, dosage_unit = ?, note = ?, occurred_at = ? WHERE id = ?",
			req.Name, req.DosageValue, req.DosageUnit, req.Note, req.OccurredAt, recordID,
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

	// 广播更新后的完整记录：家人端编辑后其他人的页面要能原位刷新。
	// recordType 已被上方 recordTables 白名单约束为非空，default 分支即 feeding。
	if rec, ok := loadSingleRecord(recordType, recordID); ok {
		BroadcastMessage(models.WebSocketMessage{
			Type:    "record_updated",
			Payload: rec,
		}, babyFamilyID(babyID))
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
	if _, ok := recordTables[recordType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须为 feeding, diaper, sleep, temperature, outdoor 或 supplement"})
		return
	}

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
	case "outdoor":
		_, err = database.DB.Exec("DELETE FROM outdoor_records WHERE id = ?", recordID)
	case "supplement":
		_, err = database.DB.Exec("DELETE FROM supplement_records WHERE id = ?", recordID)
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
	}, babyFamilyID(babyID))

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

// GetLatestTemperature 获取最近一次测温记录（用于快捷填表）
func GetLatestTemperature(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var temperature float64
	var location, note string
	err := database.DB.QueryRow(
		"SELECT temperature, location, note FROM temperature_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&temperature, &location, &note)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"temperature": temperature,
		"location":    location,
		"note":        note,
	})
}
