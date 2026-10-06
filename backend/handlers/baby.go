package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// babyFamilyID 返回宝宝所属家庭（0 = 不存在或已软删除）。广播按此投递，越界即不发。
// 注意：软删除后返回 0，需要在删除动作之前取值。
func babyFamilyID(babyID int64) int64 {
	var familyID int64
	database.DB.QueryRow(
		"SELECT u.family_id FROM babies b JOIN users u ON b.user_id = u.id WHERE b.id = ? AND b.deleted_at IS NULL",
		babyID,
	).Scan(&familyID)
	return familyID
}

// checkBabyFamily 检查宝宝是否属于当前用户的家庭（软删除的宝宝视为不可用）
func checkBabyFamily(babyID, userID int64) bool {
	familyID := babyFamilyID(babyID)
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
		WHERE u.family_id = (SELECT family_id FROM users WHERE id = ?) AND b.deleted_at IS NULL
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

// GetBaby 获取单个宝宝
func GetBaby(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || babyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var baby models.Baby
	err = database.DB.QueryRow(
		"SELECT id, user_id, name, birth_date, gender, avatar_color, created_at FROM babies WHERE id = ?",
		babyID,
	).Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.BirthDate, &baby.Gender, &baby.AvatarColor, &baby.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "宝宝不存在"})
		return
	}

	c.JSON(http.StatusOK, baby)
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
		req.AvatarColor = "#F25C8C"
	}
	req.BirthDate = normalizeBirthDate(req.BirthDate, getTzOffset(c))

	result, err := database.DB.Exec(
		"INSERT INTO babies (user_id, name, birth_date, gender, avatar_color) VALUES (?, ?, ?, ?, ?)",
		userID, req.Name, req.BirthDate, req.Gender, req.AvatarColor,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建宝宝失败"})
		return
	}

	babyID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建宝宝失败"})
		return
	}

	var baby models.Baby
	err = database.DB.QueryRow(
		"SELECT id, user_id, name, birth_date, gender, avatar_color, created_at FROM babies WHERE id = ?",
		babyID,
	).Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.BirthDate, &baby.Gender, &baby.AvatarColor, &baby.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建宝宝失败"})
		return
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "baby_created",
		Payload: baby,
	}, babyFamilyID(baby.ID))

	c.JSON(http.StatusCreated, baby)
}

// UpdateBaby 更新宝宝
func UpdateBaby(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || babyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req models.UpdateBabyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
		return
	}

	// 动态 SET：指针 DTO 下 nil 字段跳过，空串照常写入（出生日期清空、性别改保密可行）
	normalizedBirth := ""
	if req.BirthDate != nil {
		normalizedBirth = normalizeBirthDate(*req.BirthDate, getTzOffset(c))
	}
	var sets []string
	var args []interface{}
	if req.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *req.Name)
	}
	if req.BirthDate != nil {
		sets = append(sets, "birth_date = ?")
		args = append(args, normalizedBirth)
	}
	if req.Gender != nil {
		sets = append(sets, "gender = ?")
		args = append(args, *req.Gender)
	}
	if req.AvatarColor != nil {
		sets = append(sets, "avatar_color = ?")
		args = append(args, *req.AvatarColor)
	}
	if len(sets) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "无字段更新"})
		return
	}
	args = append(args, babyID)
	_, err = database.DB.Exec("UPDATE babies SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
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
	}, babyFamilyID(babyID))

	c.JSON(http.StatusOK, baby)
}

// DeleteBaby 删除宝宝
func DeleteBaby(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || babyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
		return
	}

	// 软删除：仅打 deleted_at 标记，行保留，所有记录外键不受影响。
	// familyID 必须在删除前取——babyFamilyID 忽略已软删除的行。
	familyID := babyFamilyID(babyID)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = database.DB.Exec("UPDATE babies SET deleted_at = ? WHERE id = ?", now, babyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	BroadcastMessage(models.WebSocketMessage{
		Type:    "baby_deleted",
		Payload: map[string]int64{"id": babyID},
	}, familyID)

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetStats 获取宝宝统计
func GetStats(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || babyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	tzOffset := getTzOffset(c)
	todayStart, todayEnd := todayDateRange(tzOffset)
	loc := time.FixedZone("user", tzOffset*60)
	// 今日日历日键由窗口起点推导，而非二次 time.Now()——否则恰在午夜边界时
	// 两次取时可能落在不同日期，导致 sleepToday[todayKey] 取到空值。
	todayKey := parseTime(todayStart).In(loc).Format("2006-01-02")

	var feedingCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM feeding_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ?",
		babyID, todayStart, todayEnd,
	).Scan(&feedingCount)

	var diaperCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM diaper_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ?",
		babyID, todayStart, todayEnd,
	).Scan(&diaperCount)

	var lastFeeding string
	database.DB.QueryRow(
		"SELECT occurred_at FROM feeding_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&lastFeeding)

	var lastDiaper string
	database.DB.QueryRow(
		"SELECT occurred_at FROM diaper_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&lastDiaper)

	var totalMl int
	database.DB.QueryRow(
		"SELECT COALESCE(SUM(amount_ml), 0) FROM feeding_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ? AND amount_ml > 0",
		babyID, todayStart, todayEnd,
	).Scan(&totalMl)

	sleepToday, err := sumSpansByDay("sleep_records", babyID, todayStart, todayEnd, loc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var lastSleepEnd string
	database.DB.QueryRow(
		"SELECT ended_at FROM sleep_records WHERE baby_id = ? AND ended_at IS NOT NULL ORDER BY ended_at DESC LIMIT 1",
		babyID,
	).Scan(&lastSleepEnd)

	var temperatureCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM temperature_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ?",
		babyID, todayStart, todayEnd,
	).Scan(&temperatureCount)

	var latestTemp float64
	database.DB.QueryRow(
		"SELECT COALESCE(temperature, 0) FROM temperature_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&latestTemp)

	var lastTemperature string
	database.DB.QueryRow(
		"SELECT occurred_at FROM temperature_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&lastTemperature)

	outdoorToday, err := sumSpansByDay("outdoor_records", babyID, todayStart, todayEnd, loc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var lastOutdoorEnd string
	database.DB.QueryRow(
		"SELECT ended_at FROM outdoor_records WHERE baby_id = ? AND ended_at IS NOT NULL ORDER BY ended_at DESC LIMIT 1",
		babyID,
	).Scan(&lastOutdoorEnd)

	var supplementCount int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM supplement_records WHERE baby_id = ? AND occurred_at >= ? AND occurred_at < ?",
		babyID, todayStart, todayEnd,
	).Scan(&supplementCount)

	var lastSupplement string
	database.DB.QueryRow(
		"SELECT occurred_at FROM supplement_records WHERE baby_id = ? ORDER BY occurred_at DESC LIMIT 1",
		babyID,
	).Scan(&lastSupplement)

	c.JSON(http.StatusOK, gin.H{
		"feeding_count":  feedingCount,
		"diaper_count":   diaperCount,
		"last_feeding":   lastFeeding,
		"last_diaper":    lastDiaper,
		"total_ml_today": totalMl,
		// sleep_count / outdoor_count 已删除：前端从不读取，属死字段；
		// 跨天记录该记在哪一天本身语义不明（按起始日 or 按触及日），不如不算。
		"sleep_duration":     sleepToday[todayKey],
		"last_sleep_end":     lastSleepEnd,
		"temperature_count":  temperatureCount,
		"latest_temperature": latestTemp,
		"last_temperature":   lastTemperature,
		"outdoor_duration":   outdoorToday[todayKey],
		"last_outdoor_end":   lastOutdoorEnd,
		"supplement_count":   supplementCount,
		"last_supplement":    lastSupplement,
	})
}

// DailyStats 每日统计数据结构
type DailyStats struct {
	Date            string  `json:"date"`
	FeedingCount    int     `json:"feeding_count"`
	DiaperCount     int     `json:"diaper_count"`
	TotalMl         int     `json:"total_ml"`
	SleepDuration   int     `json:"sleep_duration_minutes"`
	TemperatureAvg  float64 `json:"temperature_avg"`
	TemperatureHigh float64 `json:"temperature_high"`
	OutdoorMinutes  int     `json:"outdoor_duration_minutes"`
	SupplementCount int     `json:"supplement_count"`
}

// spanTable 白名单：只允许区间型记录表，杜绝表名拼接注入
var spanTables = map[string]bool{"sleep_records": true, "outdoor_records": true}

// sumSpansByDay 汇总区间型记录（睡眠/户外）在 [start, end) 窗口内、
// 按用户时区自然日切分后的分钟数。
//
// 为什么不用 SQL julianday 直接聚合：那样只能把整段记到 started_at 所在日，
// 跨 0 点的记录会让次日统计为 0；且 julianday 走浮点、CAST 为截断，
// 精确整数时长存在算出 599 而非 600 的边界误差。
//
// 进行中（ended_at 为空）的记录以 now 收尾，与首页「进行中」实时计时口径一致；
// 否则首页显示已睡 3h、趋势图当日为 0，两处数字互相矛盾。
//
// 注意返回值**未按窗口裁剪**：区间在窗口外但仍被重叠谓词选中的记录，
// 其窗口外的那些日期也会作为键出现（如 03-09 开始、03-10 结束的睡眠会同时给出
// 03-09 与 03-10 两项）。两个调用方都只按自己持有的日期列表取值，不受影响；
// 调用方切勿直接对返回值求和来代表「窗口内总时长」。
func sumSpansByDay(table string, babyID int64, start, end string, loc *time.Location) (map[string]int, error) {
	out := make(map[string]int)
	if !spanTables[table] {
		return out, fmt.Errorf("非法区间表: %s", table)
	}
	rows, err := database.DB.Query(
		"SELECT started_at, COALESCE(ended_at, '') FROM "+table+
			" WHERE baby_id = ?"+spanOverlapFilter(),
		babyID, end, start,
	)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	now := time.Now()
	for rows.Next() {
		var startedAt, endedAt string
		if rows.Scan(&startedAt, &endedAt) != nil {
			continue
		}
		st := parseTime(startedAt)
		if st.IsZero() {
			continue
		}
		et := now
		if endedAt != "" {
			et = parseTime(endedAt)
			if et.IsZero() {
				continue
			}
		}
		for d, mins := range splitSpanByLocalDay(st, et, loc) {
			out[d] += mins
		}
	}
	return out, rows.Err()
}

// GetTrendStats 获取宝宝趋势统计（最近7天）
func GetTrendStats(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || babyID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
		return
	}

	tzOffset := getTzOffset(c)
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}
	dates := lastNDates(tzOffset, days)
	if len(dates) == 0 {
		c.JSON(http.StatusOK, []DailyStats{})
		return
	}

	loc := time.FixedZone("user", tzOffset*60)
	windowStart, windowEnd := windowRangeUTC(tzOffset, days)

	feedingRows, err := database.DB.Query(`
		SELECT occurred_at, amount_ml FROM feeding_records
		WHERE baby_id = ? AND occurred_at >= ?
	`, babyID, windowStart)
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

	diaperRows, err := database.DB.Query(`
		SELECT occurred_at FROM diaper_records
		WHERE baby_id = ? AND occurred_at >= ?
	`, babyID, windowStart)
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

	// 睡眠 / 户外：按自然日切分时长，跨 0 点的记录两侧都计入
	sleepMap, err := sumSpansByDay("sleep_records", babyID, windowStart, windowEnd, loc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	tempRows, err := database.DB.Query(`
		SELECT occurred_at, temperature FROM temperature_records
		WHERE baby_id = ? AND occurred_at >= ?
	`, babyID, windowStart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer tempRows.Close()

	type tempEntry struct {
		avg   float64
		high  float64
		count int
	}
	tempMap := make(map[string]*tempEntry)
	for tempRows.Next() {
		var occurredAt string
		var temp float64
		if tempRows.Scan(&occurredAt, &temp) != nil {
			continue
		}
		t := parseTime(occurredAt).In(loc)
		date := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
		te, ok := tempMap[date]
		if !ok {
			te = &tempEntry{}
			tempMap[date] = te
		}
		te.count++
		te.avg += temp
		if temp > te.high {
			te.high = temp
		}
	}

	outdoorMap, err := sumSpansByDay("outdoor_records", babyID, windowStart, windowEnd, loc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	supplementRows, err := database.DB.Query(`
		SELECT occurred_at FROM supplement_records
		WHERE baby_id = ? AND occurred_at >= ?
	`, babyID, windowStart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer supplementRows.Close()

	supplementMap := make(map[string]int)
	for supplementRows.Next() {
		var occurredAt string
		if supplementRows.Scan(&occurredAt) != nil {
			continue
		}
		t := parseTime(occurredAt).In(loc)
		date := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
		supplementMap[date]++
	}

	var trends []DailyStats
	for _, date := range dates {
		f := feedingMap[date]
		var feedingCount, totalMl int
		if f != nil {
			feedingCount = f.FeedingCount
			totalMl = f.TotalMl
		}
		te := tempMap[date]
		var tempAvg, tempHigh float64
		if te != nil && te.count > 0 {
			tempAvg = te.avg / float64(te.count)
			tempHigh = te.high
		}
		trends = append(trends, DailyStats{
			Date:            date,
			FeedingCount:    feedingCount,
			DiaperCount:     diaperMap[date],
			TotalMl:         totalMl,
			SleepDuration:   sleepMap[date],
			TemperatureAvg:  tempAvg,
			TemperatureHigh: tempHigh,
			OutdoorMinutes:  outdoorMap[date],
			SupplementCount: supplementMap[date],
		})
	}

	c.JSON(http.StatusOK, trends)
}
