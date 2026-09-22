package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// --- WHO 儿童生长标准（0-24 月龄，按月龄的 L/M/S 参数）---
// 数据来源：WHO Child Growth Standards（LMS 方法）
// p = M * (1 + L*S*z)^(1/L)；L=0 时退化为 M*exp(S*z)

type lms struct{ L, M, S float64 }

// 男孩体重（kg）0-24 月，逐月
var whoWeightBoy = []lms{
	{0.3487, 3.3464, 0.14602}, {0.2297, 4.4709, 0.13395}, {0.1970, 5.5675, 0.12385},
	{0.1738, 6.3762, 0.11727}, {0.1553, 7.0023, 0.11316}, {0.1395, 7.5105, 0.11080},
	{0.1257, 7.9340, 0.10958}, {0.1134, 8.2970, 0.10902}, {0.1021, 8.6151, 0.10882},
	{0.0917, 8.9014, 0.10881}, {0.0820, 9.1649, 0.10891}, {0.0730, 9.4122, 0.10906},
	{0.0644, 9.6479, 0.10925}, {0.0563, 9.8749, 0.10949}, {0.0487, 10.0953, 0.10976},
	{0.0413, 10.3108, 0.11007}, {0.0343, 10.5228, 0.11041}, {0.0275, 10.7319, 0.11079},
	{0.0211, 10.9385, 0.11119}, {0.0148, 11.1430, 0.11164}, {0.0087, 11.3462, 0.11211},
	{0.0029, 11.5486, 0.11261}, {-0.0028, 11.7504, 0.11314}, {-0.0083, 11.9514, 0.11369},
	{-0.0137, 12.1515, 0.11426},
}

// 女孩体重（kg）0-24 月
var whoWeightGirl = []lms{
	{0.3809, 3.2322, 0.14171}, {0.1714, 4.1873, 0.13724}, {0.0962, 5.1282, 0.13000},
	{0.0402, 5.8458, 0.12619}, {-0.0050, 6.4237, 0.12402}, {-0.0430, 6.8985, 0.12274},
	{-0.0756, 7.2970, 0.12204}, {-0.1039, 7.6422, 0.12178}, {-0.1288, 7.9487, 0.12181},
	{-0.1507, 8.2254, 0.12199}, {-0.1700, 8.4800, 0.12223}, {-0.1872, 8.7192, 0.12247},
	{-0.2024, 8.9481, 0.12268}, {-0.2158, 9.1699, 0.12283}, {-0.2278, 9.3870, 0.12294},
	{-0.2384, 9.6008, 0.12299}, {-0.2478, 9.8124, 0.12303}, {-0.2562, 10.0226, 0.12306},
	{-0.2637, 10.2315, 0.12309}, {-0.2703, 10.4393, 0.12315}, {-0.2762, 10.6464, 0.12323},
	{-0.2815, 10.8534, 0.12335}, {-0.2862, 11.0608, 0.12350}, {-0.2903, 11.2688, 0.12369},
	{-0.2941, 11.4775, 0.12390},
}

// 男孩身高（cm）0-24 月
var whoHeightBoy = []lms{
	{1, 49.8842, 0.03795}, {1, 54.7244, 0.03557}, {1, 58.4249, 0.03424},
	{1, 61.4292, 0.03328}, {1, 63.8860, 0.03257}, {1, 65.9026, 0.03204},
	{1, 67.6236, 0.03165}, {1, 69.1645, 0.03139}, {1, 70.5994, 0.03124},
	{1, 71.9687, 0.03117}, {1, 73.2812, 0.03118}, {1, 74.5388, 0.03125},
	{1, 75.7488, 0.03137}, {1, 76.9186, 0.03154}, {1, 78.0497, 0.03174},
	{1, 79.1458, 0.03197}, {1, 80.2113, 0.03222}, {1, 81.2487, 0.03250},
	{1, 82.2587, 0.03279}, {1, 83.2418, 0.03310}, {1, 84.1996, 0.03342},
	{1, 85.1348, 0.03376}, {1, 86.0477, 0.03410}, {1, 86.9410, 0.03445},
	{1, 87.8161, 0.03479},
}

// 女孩身高（cm）0-24 月
var whoHeightGirl = []lms{
	{1, 49.1477, 0.03790}, {1, 53.6872, 0.03640}, {1, 57.0673, 0.03568},
	{1, 59.8029, 0.03520}, {1, 62.0899, 0.03486}, {1, 64.0301, 0.03463},
	{1, 65.7311, 0.03448}, {1, 67.2873, 0.03441}, {1, 68.7498, 0.03440},
	{1, 70.1435, 0.03444}, {1, 71.4818, 0.03452}, {1, 72.7710, 0.03464},
	{1, 74.0150, 0.03479}, {1, 75.2176, 0.03496}, {1, 76.3817, 0.03514},
	{1, 77.5099, 0.03534}, {1, 78.6055, 0.03555}, {1, 79.6710, 0.03576},
	{1, 80.7079, 0.03598}, {1, 81.7182, 0.03620}, {1, 82.7036, 0.03643},
	{1, 83.6654, 0.03666}, {1, 84.6040, 0.03688}, {1, 85.5202, 0.03711},
	{1, 86.4153, 0.03734},
}

// 男孩头围（cm）0-24 月
var whoHeadBoy = []lms{
	{1, 34.4618, 0.03686}, {1, 37.2759, 0.03133}, {1, 39.1285, 0.02997},
	{1, 40.5135, 0.02918}, {1, 41.6317, 0.02868}, {1, 42.5576, 0.02837},
	{1, 43.3306, 0.02817}, {1, 43.9803, 0.02804}, {1, 44.5300, 0.02796},
	{1, 44.9998, 0.02792}, {1, 45.4051, 0.02790}, {1, 45.7573, 0.02789},
	{1, 46.0661, 0.02789}, {1, 46.3395, 0.02789}, {1, 46.5844, 0.02791},
	{1, 46.8060, 0.02792}, {1, 47.0088, 0.02795}, {1, 47.1962, 0.02797},
	{1, 47.3711, 0.02800}, {1, 47.5357, 0.02803}, {1, 47.6919, 0.02806},
	{1, 47.8413, 0.02808}, {1, 47.9850, 0.02811}, {1, 48.1237, 0.02813},
	{1, 48.2584, 0.02815},
}

// 女孩头围（cm）0-24 月
var whoHeadGirl = []lms{
	{1, 33.8787, 0.03496}, {1, 36.5463, 0.03210}, {1, 38.2521, 0.03168},
	{1, 39.5328, 0.03140}, {1, 40.5817, 0.03119}, {1, 41.4590, 0.03102},
	{1, 42.1995, 0.03091}, {1, 42.8290, 0.03083}, {1, 43.3671, 0.03077},
	{1, 43.8300, 0.03073}, {1, 44.2319, 0.03070}, {1, 44.5844, 0.03068},
	{1, 44.8965, 0.03067}, {1, 45.1752, 0.03066}, {1, 45.4265, 0.03066},
	{1, 45.6551, 0.03066}, {1, 45.8650, 0.03067}, {1, 46.0598, 0.03068},
	{1, 46.2424, 0.03069}, {1, 46.4152, 0.03071}, {1, 46.5801, 0.03073},
	{1, 46.7388, 0.03075}, {1, 46.8925, 0.03077}, {1, 47.0423, 0.03079},
	{1, 47.1892, 0.03082},
}

// lmsAt 取月龄对应的 LMS（超出 24 月按月线性外推最近端点）
func lmsAt(table []lms, month float64) lms {
	if month < 0 {
		month = 0
	}
	if month >= float64(len(table)-1) {
		return table[len(table)-1]
	}
	i := int(month)
	f := month - float64(i)
	a, b := table[i], table[i+1]
	return lms{a.L + (b.L-a.L)*f, a.M + (b.M-a.M)*f, a.S + (b.S-a.S)*f}
}

// lmsZ LMS 法求 z 值：z = ((y/M)^L - 1) / (L*S)；L=0 时 z = ln(y/M)/S
func lmsZ(table []lms, month, value float64) float64 {
	p := lmsAt(table, month)
	if p.M == 0 || p.S == 0 {
		return 0
	}
	if math.Abs(p.L) < 1e-9 {
		return math.Log(value/p.M) / p.S
	}
	return (math.Pow(value/p.M, p.L) - 1) / (p.L * p.S)
}

// normalCDF 标准正态分布 CDF（Abramowitz-Stegun 7.1.26 近似）
func normalCDF(z float64) float64 {
	return 0.5 * (1 + math.Erf(z/math.Sqrt2))
}

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

// monthsBetween 计算月龄（按整月，月内不足一天不计）
func monthsBetween(birth, at time.Time) float64 {
	months := (at.Year()-birth.Year())*12 + int(at.Month()) - int(birth.Month())
	if at.Day() < birth.Day() {
		months--
	}
	if months < 0 {
		months = 0
	}
	return float64(months)
}

// growthPercentile 返回某项指标的百分位（0-100）与 z 值
func growthPercentile(gender string, birth time.Time, measured time.Time, metric string, value float64) (float64, float64) {
	if value <= 0 {
		return 0, 0
	}
	month := monthsBetween(birth, measured)
	var table []lms
	switch metric {
	case "weight":
		if gender == "male" {
			table = whoWeightBoy
		} else {
			table = whoWeightGirl
		}
	case "height":
		if gender == "male" {
			table = whoHeightBoy
		} else {
			table = whoHeightGirl
		}
	default: // head
		if gender == "male" {
			table = whoHeadBoy
		} else {
			table = whoHeadGirl
		}
	}
	z := lmsZ(table, month, value)
	return math.Round(normalCDF(z)*1000) / 10, math.Round(z*100) / 100
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
	AgeMonths float64 `json:"age_months"`
	Gender    string  `json:"gender"`
	WeightKg  float64 `json:"weight_kg"`
	HeightCm  float64 `json:"height_cm"`
	HeadCm    float64 `json:"head_cm"`
	WeightPct float64 `json:"weight_pct"`
	HeightPct float64 `json:"height_pct"`
	HeadPct   float64 `json:"head_pct"`
	WeightZ   float64 `json:"weight_z"`
	HeightZ   float64 `json:"height_z"`
	HeadZ     float64 `json:"head_z"`
}

// GetGrowthStats 获取最新一条成长记录的 WHO 百分位
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
	if gender != "male" {
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
	wp, wz := growthPercentile(gender, birth, measured, "weight", g.WeightKg)
	hp, hz := growthPercentile(gender, birth, measured, "height", g.HeightCm)
	cp, cz := growthPercentile(gender, birth, measured, "head", g.HeadCm)

	c.JSON(http.StatusOK, GrowthStats{
		AgeMonths: monthsBetween(birth, measured),
		Gender:    gender,
		WeightKg:  g.WeightKg, HeightCm: g.HeightCm, HeadCm: g.HeadCm,
		WeightPct: wp, HeightPct: hp, HeadPct: cp,
		WeightZ: wz, HeightZ: hz, HeadZ: cz,
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
	if req.WeightKg < 0 || req.HeightCm < 0 || req.HeadCm < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数值不能为负"})
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
	BroadcastMessage(models.WebSocketMessage{Type: "record_created", Payload: gin.H{"id": id, "record_type": "growth"}})
	c.JSON(http.StatusOK, gin.H{"id": id})
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
	database.DB.QueryRow("SELECT baby_id FROM growth_records WHERE id = ?", id).Scan(&babyID)
	if babyID == 0 || !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	database.DB.Exec("DELETE FROM growth_records WHERE id = ?", id)
	BroadcastMessage(models.WebSocketMessage{Type: "record_deleted", Payload: gin.H{"id": id, "record_type": "growth"}})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
