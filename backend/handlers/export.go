package handlers

import (
	"baby-care-tracker/database"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// csvRow CSV 导出的一行
type csvRow struct {
	t      time.Time
	kind   string
	detail string
	note   string
}

// ExportRecords 导出某宝宝的全部记录为 CSV（带 UTF-8 BOM，Excel 直接可读）
// GET /api/babies/:id/export?days=7
func ExportRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	babyID, ok := parseID(c)
	if !ok {
		return
	}
	if !checkBabyFamily(babyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	var babyName string
	database.DB.QueryRow("SELECT name FROM babies WHERE id = ?", babyID).Scan(&babyName)
	if babyName == "" {
		babyName = "baby"
	}

	tzOffset := getTzOffset(c)

	// 时间窗口（可选）
	startStr := ""
	if ds := c.Query("days"); ds != "" {
		if days, err := strconv.Atoi(ds); err == nil && days > 0 && days <= 365 {
			startStr = daysAgoUTC(tzOffset, days)
		}
	}
	occurredFilter, occurredArgs := "", []interface{}{babyID}
	startedFilter, startedArgs := "", []interface{}{babyID}
	if startStr != "" {
		occurredFilter = " AND occurred_at >= ?"
		occurredArgs = append(occurredArgs, startStr)
		startedFilter = " AND started_at >= ?"
		startedArgs = append(startedArgs, startStr)
	}

	var rows []csvRow

	// 喂奶
	if rs, err := database.DB.Query(
		"SELECT type, duration_minutes, amount_ml, side, brand, note, occurred_at FROM feeding_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var typ, side, brand, note, occ string
			var dur, amt int
			rs.Scan(&typ, &dur, &amt, &side, &brand, &note, &occ)
			detail := map[string]string{"breast": "母乳", "bottle": "瓶喂", "formula": "配方奶"}[typ]
			if detail == "" {
				detail = typ
			}
			if amt > 0 {
				detail += fmt.Sprintf(" %dml", amt)
			}
			if dur > 0 {
				detail += fmt.Sprintf(" %d分钟", dur)
			}
			if side != "" {
				detail += fmt.Sprintf(" (%s)", map[string]string{"left": "左", "right": "右", "both": "双侧"}[side])
			}
			if brand != "" {
				detail += " " + brand
			}
			rows = append(rows, csvRow{parseTime(occ), "喂奶", detail, note})
		}
		rs.Close()
	}

	// 尿布
	if rs, err := database.DB.Query(
		"SELECT type, note, occurred_at FROM diaper_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var typ, note, occ string
			rs.Scan(&typ, &note, &occ)
			detail := map[string]string{"pee": "小便", "poop": "大便", "mixed": "混合"}[typ]
			if detail == "" {
				detail = typ
			}
			rows = append(rows, csvRow{parseTime(occ), "尿布", detail, note})
		}
		rs.Close()
	}

	// 睡眠
	if rs, err := database.DB.Query(
		"SELECT started_at, ended_at, note FROM sleep_records WHERE baby_id = ? AND ended_at IS NOT NULL"+startedFilter+" ORDER BY started_at DESC",
		startedArgs...,
	); err == nil {
		for rs.Next() {
			var st, en, note string
			rs.Scan(&st, &en, &note)
			detail := ""
			if t1, t2 := parseTime(st), parseTime(en); !t1.IsZero() && !t2.IsZero() {
				detail = fmt.Sprintf("%.0f分钟", t2.Sub(t1).Minutes())
			}
			rows = append(rows, csvRow{parseTime(st), "睡眠", detail, note})
		}
		rs.Close()
	}

	// 体温
	if rs, err := database.DB.Query(
		"SELECT temperature, location, note, occurred_at FROM temperature_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var temp float64
			var loc, note, occ string
			rs.Scan(&temp, &loc, &note, &occ)
			detail := fmt.Sprintf("%.1f°C", temp)
			if loc != "" {
				detail += " " + loc
			}
			rows = append(rows, csvRow{parseTime(occ), "体温", detail, note})
		}
		rs.Close()
	}

	// 户外
	if rs, err := database.DB.Query(
		"SELECT started_at, ended_at, note FROM outdoor_records WHERE baby_id = ? AND ended_at IS NOT NULL"+startedFilter+" ORDER BY started_at DESC",
		startedArgs...,
	); err == nil {
		for rs.Next() {
			var st, en, note string
			rs.Scan(&st, &en, &note)
			detail := ""
			if t1, t2 := parseTime(st), parseTime(en); !t1.IsZero() && !t2.IsZero() {
				detail = fmt.Sprintf("%.0f分钟", t2.Sub(t1).Minutes())
			}
			rows = append(rows, csvRow{parseTime(st), "户外", detail, note})
		}
		rs.Close()
	}

	// 补剂
	if rs, err := database.DB.Query(
		"SELECT name, dosage_value, dosage_unit, note, occurred_at FROM supplement_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var name, unit, note, occ string
			var val float64
			rs.Scan(&name, &val, &unit, &note, &occ)
			detail := name
			if val > 0 {
				detail += fmt.Sprintf(" %.1f%s", val, unit)
			}
			rows = append(rows, csvRow{parseTime(occ), "补剂", detail, note})
		}
		rs.Close()
	}

	// 成长（身高/体重/头围）
	if rs, err := database.DB.Query(
		"SELECT measured_at, weight_kg, height_cm, head_cm, note FROM growth_records WHERE baby_id = ? ORDER BY measured_at DESC",
		babyID,
	); err == nil {
		for rs.Next() {
			var occ, note string
			var w, h, hd float64
			rs.Scan(&occ, &w, &h, &hd, &note)
			var parts []string
			if w > 0 {
				parts = append(parts, fmt.Sprintf("体重 %.2fkg", w))
			}
			if h > 0 {
				parts = append(parts, fmt.Sprintf("身高 %.1fcm", h))
			}
			if hd > 0 {
				parts = append(parts, fmt.Sprintf("头围 %.1fcm", hd))
			}
			detail := strings.Join(parts, " ")
			rows = append(rows, csvRow{measuredTimeFromDB(occ), "成长", detail, note})
		}
		rs.Close()
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].t.After(rows[j].t) })

	loc := time.FixedZone("user", tzOffset*60)
	filename := fmt.Sprintf("%s-%s.csv", babyName, time.Now().In(loc).Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	// RFC 5987 编码，兼容中文文件名
	c.Header("Content-Disposition", "attachment; filename=\"records.csv\"; filename*=UTF-8''"+url.PathEscape(filename))

	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"时间", "类型", "详情", "备注"})
	for _, r := range rows {
		w.Write([]string{
			r.t.In(loc).Format("2006-01-02 15:04"),
			r.kind,
			r.detail,
			r.note,
		})
	}
	w.Flush()
}
