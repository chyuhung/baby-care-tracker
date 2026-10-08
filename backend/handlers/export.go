package handlers

import (
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

// guardCSVCell 防 CSV 公式注入：单元格以 = + - @ TAB CR 开头时，Excel/WPS 会把它当公式执行
// （=HYPERLINK(...)、+cmd|... 等）。前置一个单引号使其按纯文本显示，转义只作用于首个危险字符。
func guardCSVCell(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}

// spanDetail 生成区间型记录（睡眠/户外）的明细文案。
// 进行中（en 为空）输出「进行中」而非空白或 0 分钟，避免 CSV 里出现
// 一行没有时长却也没标明状态的睡眠记录。
func spanDetail(start, end string) string {
	if end == "" {
		return "进行中"
	}
	t1, t2 := parseTime(start), parseTime(end)
	if t1.IsZero() || t2.IsZero() {
		return ""
	}
	return fmt.Sprintf("%.0f分钟", t2.Sub(t1).Minutes())
}

// csvRow CSV 导出的一行
type csvRow struct {
	t      time.Time
	kind   string
	detail string
	note   string
}

// csvRows 把共享取数层的数据渲染成 CSV 行（新→旧，全类型按时间合并）。
// detail 文案与历史实现逐字一致，防止导出回归。
func (d *reportData) csvRows() []csvRow {
	var rows []csvRow
	for _, f := range d.Feeding {
		detail := map[string]string{"breast": "母乳", "bottle": "瓶喂", "formula": "配方奶"}[f.Type]
		if detail == "" {
			detail = f.Type
		}
		if f.Amt > 0 {
			detail += fmt.Sprintf(" %dml", f.Amt)
		}
		if f.Dur > 0 {
			detail += fmt.Sprintf(" %d分钟", f.Dur)
		}
		if f.Side != "" {
			detail += fmt.Sprintf(" (%s)", map[string]string{"left": "左", "right": "右", "both": "双侧"}[f.Side])
		}
		if f.Brand != "" {
			detail += " " + f.Brand
		}
		rows = append(rows, csvRow{f.T, "喂奶", detail, f.Note})
	}
	for _, x := range d.Diaper {
		detail := map[string]string{"pee": "小便", "poop": "大便", "mixed": "混合"}[x.Type]
		if detail == "" {
			detail = x.Type
		}
		rows = append(rows, csvRow{x.T, "尿布", detail, x.Note})
	}
	for _, s := range d.Sleep {
		rows = append(rows, csvRow{s.Start, "睡眠", spanOf(s), s.Note})
	}
	for _, t := range d.Temp {
		detail := fmt.Sprintf("%.1f°C", t.Val)
		if t.Loc != "" {
			detail += " " + t.Loc
		}
		rows = append(rows, csvRow{t.T, "体温", detail, t.Note})
	}
	for _, s := range d.Outdoor {
		rows = append(rows, csvRow{s.Start, "户外", spanOf(s), s.Note})
	}
	for _, su := range d.Supplement {
		detail := su.Name
		if su.Val > 0 {
			detail += fmt.Sprintf(" %.1f%s", su.Val, su.Unit)
		}
		rows = append(rows, csvRow{su.T, "补剂", detail, su.Note})
	}
	for _, g := range d.Growth {
		var parts []string
		if g.H > 0 {
			parts = append(parts, fmt.Sprintf("身高 %.1fcm", g.H))
		}
		if g.W > 0 {
			parts = append(parts, fmt.Sprintf("体重 %.2fkg", g.W))
		}
		if g.Hd > 0 {
			parts = append(parts, fmt.Sprintf("头围 %.1fcm", g.Hd))
		}
		rows = append(rows, csvRow{g.T, "成长", strings.Join(parts, " "), g.Note})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].t.After(rows[j].t) })
	return rows
}

// spanOf 区间记录（睡眠/户外）的明细文案；进行中输出「进行中」而非空白/0 分钟。
func spanOf(s spanRow) string {
	if s.Ongoing {
		return "进行中"
	}
	return spanDetail(s.Start.Format(time.RFC3339), s.End.Format(time.RFC3339))
}

// ExportRecords 导出某宝宝的全部记录。
// GET /api/babies/:id/export?days=7&format=csv|pdf（默认 csv）
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

	tzOffset := getTzOffset(c)

	// 时间窗口（可选）
	days := 0
	if ds := c.Query("days"); ds != "" {
		if v, err := strconv.Atoi(ds); err == nil && v > 0 && v <= 365 {
			days = v
		}
	}

	if c.Query("format") == "pdf" {
		writePDF(c, babyID, userID, tzOffset, days)
		return
	}

	d := queryReportData(babyID, userID, tzOffset, days)
	babyName := d.BabyName
	if babyName == "" {
		babyName = "baby"
	}
	rows := d.csvRows()

	loc := time.FixedZone("user", tzOffset*60)
	filename := fmt.Sprintf("%s-%s.csv", babyName, time.Now().In(loc).Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	// RFC 5987 编码，兼容中文文件名
	c.Header("Content-Disposition", "attachment; filename=\"records.csv\"; filename*=UTF-8''"+url.PathEscape(filename))

	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"时间", "类型", "详情", "备注"})
	for _, r := range rows {
		// 成长记录的测量日期是纯日历日（YYYY-MM-DD），不做时区换算；
		// 其余记录为时刻，转用户时区显示
		val := r.t.Format("2006-01-02")
		if r.kind != "成长" {
			val = r.t.In(loc).Format("2006-01-02 15:04")
		}
		// detail 与 note 为用户可控内容，过公式注入防护
		w.Write([]string{val, r.kind, guardCSVCell(r.detail), guardCSVCell(r.note)})
	}
	w.Flush()
}
