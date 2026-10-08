package handlers

import (
	"baby-care-tracker/database"
	growthdata "baby-care-tracker/data"
	"baby-care-tracker/models"
	"database/sql"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// validateGrowthMetrics 返回空串表示通过。上限拦截 999 这类误输——它们会把曲线 Y 轴撑到上千，
// 且明显不可能是婴幼儿的测量值。零值字段（未填）不算错误。
func validateGrowthMetrics(weightKg, heightCm, headCm float64) string {
	if weightKg < 0 || weightKg > 60 {
		return "体重需在 0-60kg 之间"
	}
	if heightCm < 0 || heightCm > 200 {
		return "身高需在 0-200cm 之间"
	}
	if headCm < 0 || headCm > 100 {
		return "头围需在 0-100cm 之间"
	}
	return ""
}

// --- 儿童生长百分位（《7岁以下儿童生长标准》WS/T 423-2022）---
// 数据见 backend/data（内嵌 JSON 百分位表，0-11 月逐月、其后每 3 月一行，
// 头围 0-3 岁）。百分位由 P3/P10/P25/P50/P75/P90/P97 相邻点分段线性插值。

// birthTimeFromDB 解析 babies.birth_date（通常为 RFC3339，如 2026-06-01T08:00:00Z）
func birthTimeFromDB(s string) time.Time {
	if len(s) >= 10 {
		if t, err := time.Parse("2006-01-02", s[:10]); err == nil {
			return t
		}
	}
	return parseTime(s)
}

// measuredTimeFromDB 解析 growth_records.measured_at（YYYY-MM-DD）
func measuredTimeFromDB(s string) time.Time {
	if len(s) >= 10 {
		if t, err := time.Parse("2006-01-02", s[:10]); err == nil {
			return t
		}
	}
	return parseTime(s)
}

// monthsBetween 计算月龄（按整月，月内不足一天不计）。
//
// birth 为零值（宝宝未填出生日期）时返回 -1 表示「月龄未知」：
// 不能返回 0——零值 birth 会算出 (measured.Year-1)*12 这种上万月的假月龄，
// 被 GrowthStats.AgeMonths 显示成「1200 月龄」，并让标准判定钳制到表末行而误判。
func monthsBetween(birth, at time.Time) float64 {
	if birth.IsZero() {
		return -1
	}
	months := (at.Year()-birth.Year())*12 + int(at.Month()) - int(birth.Month())
	if at.Day() < birth.Day() {
		months--
	}
	if months < 0 {
		months = 0
	}
	return float64(months)
}

// growthPercentile 返回某项指标的连续百分位（0-100），由官方七点分段线性插值得到。
// 注意：表1 只定义五级评价、不定义连续百分位换算，故本值仅供展示参考，
// 正式判定请用 growthGrade（见 WS/T 423-2022 表1）。
// 月龄未知时返回 0；月龄超出标准表覆盖范围时由 Percentile 自行拦截。
func growthPercentile(gender string, birth time.Time, measured time.Time, metric string, value float64) float64 {
	if value <= 0 {
		return 0
	}
	month := monthsBetween(birth, measured)
	if month < 0 {
		return 0
	}
	return math.Round(growthdata.Percentile(metric, gender, month, value)*10) / 10
}

// growthGrade 返回某项指标的五级评价（上/中上/中/中下/下），
// 依据 WS/T 423-2022 表1「儿童生长水平的百分位数评价方法」。
// 正常范围 P3–P97；仅 <P3（下）与 ≥P97（上）落在正常范围之外。
// 月龄未知或月龄超出标准表覆盖范围时返回空串（界面显示「--」）。
func growthGrade(gender string, birth time.Time, measured time.Time, metric string, value float64) string {
	if value <= 0 {
		return ""
	}
	return growthdata.Grade(metric, gender, monthsBetween(birth, measured), value)
}

// GetGrowthRecords 获取成长记录
func GetGrowthRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	rows, err := database.DB.Query(
		`SELECT id, baby_id, user_id, measured_at, weight_kg, height_cm, head_cm, note, created_at
		FROM growth_records WHERE baby_id = ? ORDER BY measured_at ASC`,
		babyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var list []models.GrowthRecord
	for rows.Next() {
		var g models.GrowthRecord
		if err := rows.Scan(&g.ID, &g.BabyID, &g.UserID, &g.MeasuredAt, &g.WeightKg, &g.HeightCm, &g.HeadCm, &g.Note, &g.CreatedAt); err != nil {
			continue
		}
		list = append(list, g)
	}
	if list == nil {
		list = []models.GrowthRecord{}
	}
	c.JSON(http.StatusOK, list)
}

// GrowthStats 成长百分位响应
type GrowthStats struct {
	// AgeMonths 为测量时的月龄（整月，日内不足一天不计）。
	// 为 -1 表示宝宝未填出生日期、月龄未知，此时前端不显示月龄，
	// 且各 *_grade 均为空串（标准判定需要月龄）。
	AgeMonths float64 `json:"age_months"`
	Gender    string  `json:"gender"`
	// GenderLabel 为档案里实际选择的中文性别（保密则为「保密」，不会被静默改写成女宝）
	GenderLabel string `json:"gender_label"`
	// GenderFallback 为 true 时表示档案未选择性别（保密），百分位/参考曲线暂按女宝标准计算
	GenderFallback bool    `json:"gender_fallback"`
	WeightKg       float64 `json:"weight_kg"`
	HeightCm       float64 `json:"height_cm"`
	HeadCm         float64 `json:"head_cm"`
	WeightPct      float64 `json:"weight_pct"`
	HeightPct      float64 `json:"height_pct"`
	HeadPct        float64 `json:"head_pct"`
	// 五级评价（WS/T 423-2022 表1）：上/中上/中/中下/下；
	// 空串表示「无档位」——可能是该项未测量、月龄未知，或月龄超出该指标表覆盖范围
	// （头围标准仅至 36 月、体重身高至 81 月）
	WeightGrade string `json:"weight_grade"`
	HeightGrade string `json:"height_grade"`
	HeadGrade   string `json:"head_grade"`
}

// GetGrowthStats 获取最新一条成长记录的标准百分位（WS/T 423-2022）
// GET /api/babies/:id/growth/stats
func GetGrowthStats(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var gender, birthDate string
	if err := database.DB.QueryRow("SELECT gender, birth_date FROM babies WHERE id = ?", babyID).Scan(&gender, &birthDate); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "宝宝不存在"})
		return
	}
	// 标准表只有男/女两套；档案未选择性别（保密）时暂按女宝标准计算，
	// 但 gender_label 仍如实回传空串，避免前端把它显示成女宝
	genderLabel := gender
	if gender != "male" && gender != "female" {
		genderLabel = ""
		gender = "female"
	}
	birth := birthTimeFromDB(birthDate)

	var g models.GrowthRecord
	err := database.DB.QueryRow(
		`SELECT id, baby_id, user_id, measured_at, weight_kg, height_cm, head_cm, note, created_at
		FROM growth_records WHERE baby_id = ? ORDER BY measured_at DESC LIMIT 1`,
		babyID,
	).Scan(&g.ID, &g.BabyID, &g.UserID, &g.MeasuredAt, &g.WeightKg, &g.HeightCm, &g.HeadCm, &g.Note, &g.CreatedAt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"empty": true})
		return
	}

	measured := measuredTimeFromDB(g.MeasuredAt)
	wp := growthPercentile(gender, birth, measured, "weight", g.WeightKg)
	hp := growthPercentile(gender, birth, measured, "height", g.HeightCm)
	cp := growthPercentile(gender, birth, measured, "head", g.HeadCm)

	c.JSON(http.StatusOK, GrowthStats{
		AgeMonths:      monthsBetween(birth, measured),
		Gender:         gender,
		GenderLabel:    genderLabel,
		GenderFallback: genderLabel == "",
		WeightKg:       g.WeightKg, HeightCm: g.HeightCm, HeadCm: g.HeadCm,
		WeightPct: wp, HeightPct: hp, HeadPct: cp,
		WeightGrade: growthGrade(gender, birth, measured, "weight", g.WeightKg),
		HeightGrade: growthGrade(gender, birth, measured, "height", g.HeightCm),
		HeadGrade:   growthGrade(gender, birth, measured, "head", g.HeadCm),
	})
}

// GrowthReferencePoint 参考曲线上一个月的五点值
type GrowthReferencePoint struct {
	Month float64 `json:"month"`
	P3    float64 `json:"p3"`
	P25   float64 `json:"p25"`
	P50   float64 `json:"p50"`
	P75   float64 `json:"p75"`
	P97   float64 `json:"p97"`
}

// GrowthReferenceMetric 一项指标的参考曲线
type GrowthReferenceMetric struct {
	Unit      string                 `json:"unit"`
	MaxMonths int                    `json:"max_months"`
	Points    []GrowthReferencePoint `json:"points"`
}

// GrowthReference 成长参考曲线响应（按宝宝性别）
type GrowthReference struct {
	Gender string                `json:"gender"`
	Weight GrowthReferenceMetric `json:"weight"`
	Height GrowthReferenceMetric `json:"height"`
	Head   GrowthReferenceMetric `json:"head"`
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func buildReferenceMetric(metric, sex, unit string) GrowthReferenceMetric {
	rows := growthdata.Table(metric, sex)
	if len(rows) == 0 {
		// 内嵌标准表缺失时防御性返回空曲线，避免 rows[len-1] 越界 panic（理论上不可达）
		return GrowthReferenceMetric{Unit: unit}
	}
	max := len(rows) - 1
	points := make([]GrowthReferencePoint, 0, int(rows[max].M)+1)
	for m := 0; m <= int(rows[max].M); m++ {
		mo := float64(m)
		points = append(points, GrowthReferencePoint{
			Month: mo,
			P3:    round1(growthdata.Value(metric, sex, "p3", mo)),
			P25:   round1(growthdata.Value(metric, sex, "p25", mo)),
			P50:   round1(growthdata.Value(metric, sex, "p50", mo)),
			P75:   round1(growthdata.Value(metric, sex, "p75", mo)),
			P97:   round1(growthdata.Value(metric, sex, "p97", mo)),
		})
	}
	return GrowthReferenceMetric{Unit: unit, MaxMonths: int(rows[max].M), Points: points}
}

// GetGrowthReference 返回宝宝性别的标准参考曲线（P3/P25/P50/P75/P97，逐月）
// GET /api/babies/:id/growth/reference
func GetGrowthReference(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var gender string
	if err := database.DB.QueryRow("SELECT gender FROM babies WHERE id = ?", babyID).Scan(&gender); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "宝宝不存在"})
		return
	}
	if gender != "male" {
		gender = "female"
	}

	c.JSON(http.StatusOK, GrowthReference{
		Gender: gender,
		Weight: buildReferenceMetric("weight", gender, "kg"),
		Height: buildReferenceMetric("height", gender, "cm"),
		Head:   buildReferenceMetric("head", gender, "cm"),
	})
}

// CreateGrowthRecord 新增成长记录
func CreateGrowthRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	var req models.CreateGrowthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "测量日期必填"})
		return
	}
	if msg := validateGrowthMetrics(req.WeightKg, req.HeightCm, req.HeadCm); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	res, err := database.DB.Exec(
		"INSERT INTO growth_records (baby_id, user_id, measured_at, weight_kg, height_cm, head_cm, note) VALUES (?, ?, ?, ?, ?, ?, ?)",
		babyID, userID, req.MeasuredAt, req.WeightKg, req.HeightCm, req.HeadCm, req.Note,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	id, _ := res.LastInsertId()
	BroadcastMessage(models.WebSocketMessage{Type: "record_created", Payload: gin.H{"id": id, "type": "growth", "baby_id": babyID}}, babyFamilyID(babyID))
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateGrowthRecord 更新成长记录
func UpdateGrowthRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseInt64(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	var babyID int64
	if err := database.DB.QueryRow("SELECT baby_id FROM growth_records WHERE id = ?", id).Scan(&babyID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	var req models.CreateGrowthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "测量日期必填"})
		return
	}
	if msg := validateGrowthMetrics(req.WeightKg, req.HeightCm, req.HeadCm); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if _, err := database.DB.Exec(
		"UPDATE growth_records SET measured_at = ?, weight_kg = ?, height_cm = ?, head_cm = ?, note = ? WHERE id = ?",
		req.MeasuredAt, req.WeightKg, req.HeightCm, req.HeadCm, req.Note, id,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	BroadcastMessage(models.WebSocketMessage{Type: "record_updated", Payload: gin.H{"id": id, "type": "growth", "baby_id": babyID}}, babyFamilyID(babyID))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DeleteGrowthRecord 删除成长记录
func DeleteGrowthRecord(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseInt64(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	var babyID int64
	if err := database.DB.QueryRow("SELECT baby_id FROM growth_records WHERE id = ?", id).Scan(&babyID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	if _, err := database.DB.Exec("DELETE FROM growth_records WHERE id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	BroadcastMessage(models.WebSocketMessage{Type: "record_deleted", Payload: gin.H{"id": id, "type": "growth", "baby_id": babyID}}, babyFamilyID(babyID))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
