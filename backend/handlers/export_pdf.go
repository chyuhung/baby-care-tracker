package handlers

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

//go:embed fonts/DroidSansFallbackFull.ttf
var reportCJKFont []byte

// ---- PDF 数据报表 ----
// 版面对齐 App 设计语言：粉色品牌强调、0.5px hairline、iOS 系统灰文本层级。
// 字体用 Droid Sans Fallback（TrueType，Apache 2.0）：go-pdf/fpdf 的 TTF 解析器
// 只认 TrueType glyf 轮廓（Noto Sans SC 是 OTF/CFF，无法内嵌），故选它。

const (
	pdfMarginX = 14.0
	pdfBottom  = 20.0 // 页脚区高度（Also 自动分页的下边界）
)

// PDF 配色
var (
	pdfInk     = [3]int{28, 28, 30}
	pdfSub     = [3]int{142, 142, 147}
	pdfHair    = [3]int{229, 229, 234}
	pdfAcc     = [3]int{255, 123, 164}
	pdfAccDeep = [3]int{215, 60, 120}
	pdfTile    = [3]int{255, 246, 250}
	pdfRowAlt  = [3]int{247, 247, 252}
	pdfHead    = [3]int{242, 242, 247}
	pdfSuccess = [3]int{52, 199, 89}
	pdfDanger  = [3]int{255, 69, 58}
)

// writePDF 生成并写出 PDF 报表。
func writePDF(c *gin.Context, babyID, userID int64, tzOffset, days int) {
	d := queryReportData(babyID, userID, tzOffset, days)
	buf, err := buildPDF(d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 PDF 失败"})
		return
	}
	name := d.BabyName
	if name == "" {
		name = "baby"
	}
	filename := fmt.Sprintf("%s-%s.pdf", name, d.Generated.Format("20060102"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\"records.pdf\"; filename*=UTF-8''"+url.PathEscape(filename))
	c.Writer.Write(buf)
}

func buildPDF(d *reportData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMarginX, 16, pdfMarginX)
	pdf.SetAutoPageBreak(true, pdfBottom)
	pdf.AddUTF8FontFromBytes("droid", "", reportCJKFont)
	pageW, pageH := pdf.GetPageSize()
	name := d.BabyName
	if name == "" {
		name = "宝宝"
	}
	genText := d.Generated.Format("2006-01-02 15:04")
	loc := d.Generated.Location()

	// SetHeaderFuncMode(..., true)：AddPage/自动分页调用 header 后把当前点重置回
	// 左上边距（tMargin），否则 header 里 SetXY(...,8) 会把 Y 留在 8mm——低于
	// tMargin(16mm)，紧接的内容会压到页眉线上。
	// v0.9.0 没有 SetFooterFuncMode（仅 SetHeaderFuncMode），但 footer 是在旧页
	// endpage 之前调用，新页 Y 由 beginpage 置为 tMargin，故 footer 不受影响；
	// footer 内部一律用相对底部的负 Y，不外泄当前点。
	pdf.SetHeaderFuncMode(func() {
		pdf.SetFont("droid", "", 8.5)
		pdf.SetTextColor(pdfAccDeep[0], pdfAccDeep[1], pdfAccDeep[2])
		pdf.SetXY(pdfMarginX, 8)
		pdf.CellFormat(70, 5, "宝宝护理记录", "", 0, "L", false, 0, "")
		pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])
		pdf.CellFormat(0, 5, "数据报告 · "+name, "", 0, "R", false, 0, "")
		pdf.SetDrawColor(pdfHair[0], pdfHair[1], pdfHair[2])
		pdf.SetLineWidth(0.2)
		pdf.Line(pdfMarginX, 14.5, pageW-pdfMarginX, 14.5)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-16)
		pdf.SetDrawColor(pdfHair[0], pdfHair[1], pdfHair[2])
		pdf.SetLineWidth(0.2)
		pdf.Line(pdfMarginX, pdf.GetY(), pageW-pdfMarginX, pdf.GetY())
		pdf.SetY(-13.5)
		pdf.SetFont("droid", "", 7.5)
		pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])
		pdf.CellFormat(0, 4, fmt.Sprintf("导出时间 %s · 第 %d / {nb} 页", genText, pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	// 必须在首个 AddPage 之前注册，否则页脚里的 {nb} 不会被替换成总页数（原样输出）。
	pdf.AliasNbPages("")

	pdf.AddPage()

	// 标题块：宝宝名 + 性别 pill + 副标题。整块显式跟踪 y，算出块底后再交给
	// 下一块，避免靠 Cell/Ln 的隐式推进（w=0 的 Cell 会把 X 顶到右边距，pill
	// 会被推出页外）与 drawBabyCard 重叠。
	titleY := pdf.GetY() + 2
	titleH := 16.0
	pdf.SetFont("droid", "", 20)
	pdf.SetTextColor(pdfInk[0], pdfInk[1], pdfInk[2])
	nameW := pdf.GetStringWidth(name)
	pdf.SetXY(pdfMarginX, titleY)
	pdf.CellFormat(nameW, 10, name, "", 0, "L", false, 0, "")
	// 性别 pill 紧贴名称右侧
	pdf.SetFont("droid", "", 8)
	pdf.SetFillColor(pdfAcc[0], pdfAcc[1], pdfAcc[2])
	pdf.SetTextColor(255, 255, 255)
	gw := pdf.GetStringWidth(genderText(d.Gender)) + 5
	pdf.SetXY(pdfMarginX+nameW+3, titleY+2)
	pdf.CellFormat(gw, 6, genderText(d.Gender), "", 0, "C", true, 0, "")
	// 副标题
	pdf.SetFont("droid", "", 9)
	pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])
	pdf.SetXY(pdfMarginX, titleY+11)
	pdf.CellFormat(0, 5, "数据报告 · 全面记录宝宝日常", "", 0, "L", false, 0, "")
	pdf.SetY(titleY + titleH)

	drawBabyCard(pdf, d, pageW, pageH)
	pdf.Ln(5)
	drawOverviewTiles(pdf, d, pageW, pageH)
	pdf.Ln(3)

	// 七个明细分区
	writeSection(pdf, pageW, pageH, "喂奶记录", len(d.Feeding), colFeed, feedingRows(d, loc))
	writeSection(pdf, pageW, pageH, "尿布记录", len(d.Diaper), colDia, diaperRows(d, loc))
	writeSection(pdf, pageW, pageH, "睡眠记录", len(d.Sleep), colSpan, spanRows(d.Sleep, loc))
	writeSection(pdf, pageW, pageH, "体温记录", len(d.Temp), colTemp, tempRows(d, loc))
	writeSection(pdf, pageW, pageH, "户外活动记录", len(d.Outdoor), colSpan, spanRows(d.Outdoor, loc))
	writeSection(pdf, pageW, pageH, "补剂记录", len(d.Supplement), colSupp, suppRows(d, loc))
	writeSection(pdf, pageW, pageH, "成长记录", len(d.Growth), colGrow, growthRows(d))

	drawGrowthCompare(pdf, d, pageW, pageH)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ---- 版式基元 ----

func ensureSpace(pdf *fpdf.Fpdf, h, pageH float64) {
	if pdf.GetY()+h > pageH-pdfBottom {
		pdf.AddPage()
	}
}

type pdfCol struct {
	header string
	width  float64
}

// 各分区的表格列（宽度总和必须等于 A4 内容宽 182mm）
var (
	colFeed = []pdfCol{{"时间", 40}, {"类型", 18}, {"奶量", 18}, {"时长", 18}, {"备注", 88}}
	colDia  = []pdfCol{{"时间", 40}, {"类型", 20}, {"备注", 122}}
	colSpan = []pdfCol{{"开始", 40}, {"结束", 36}, {"时长", 22}, {"备注", 84}}
	colTemp = []pdfCol{{"时间", 40}, {"温度", 20}, {"部位", 26}, {"备注", 96}}
	colSupp = []pdfCol{{"时间", 40}, {"名称", 46}, {"剂量", 24}, {"备注", 72}}
	colGrow = []pdfCol{{"日期", 38}, {"月龄", 14}, {"身高", 24}, {"体重", 24}, {"头围", 24}, {"备注", 58}}
)

func writeSection(pdf *fpdf.Fpdf, pageW, pageH float64, title string, count int, cols []pdfCol, rows [][]string) {
	ensureSpace(pdf, 22, pageH)
	y := pdf.GetY()
	pdf.SetFillColor(pdfAcc[0], pdfAcc[1], pdfAcc[2])
	pdf.Rect(pdfMarginX, y+0.5, 1.6, 6.6, "F")
	pdf.SetFont("droid", "", 12.5)
	pdf.SetTextColor(pdfInk[0], pdfInk[1], pdfInk[2])
	pdf.SetXY(pdfMarginX+4, y)
	pdf.CellFormat(90, 7.6, title, "", 0, "L", false, 0, "")
	pdf.SetFont("droid", "", 8)
	pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])
	pdf.CellFormat(40, 7.6, fmt.Sprintf("共 %d 条", count), "", 0, "L", false, 0, "")
	pdf.SetY(y + 10)
	drawTable(pdf, pageW, pageH, cols, rows)
	pdf.Ln(3)
}

// drawTable 画一张带表头重复的表格：交替浅灰行 + hairline 网格，跨页自动
// AddPage + 重绘表头并重置斑马纹。
//
// 单元格渲染不用 MultiCell：MultiCell 会在内部推进 Y，多列同排时相互干扰。
// 改为先按列用 SplitText 拆成行数组，再逐行 SetXY + CellFormat 精确落点。
func drawTable(pdf *fpdf.Fpdf, pageW, pageH float64, cols []pdfCol, rows [][]string) {
	tableW := pageW - 2*pdfMarginX
	const (
		cellFont = 8.5
		padX     = 1.5
		padY     = 0.8
		headerH  = 6.0
	)
	// 行高取「原基准」与「字体自然行高」的较大者（PointConvert 把 pt 换成 mm）。
	lineH := 4.2
	if fh := pdf.PointConvert(cellFont * 1.2); fh > lineH {
		lineH = fh
	}

	drawHeader := func() {
		if pdf.GetY()+headerH+2 > pageH-pdfBottom {
			pdf.AddPage()
		}
		y0 := pdf.GetY()
		pdf.SetFillColor(pdfHead[0], pdfHead[1], pdfHead[2])
		pdf.Rect(pdfMarginX, y0, tableW, headerH, "F")
		x := pdfMarginX
		pdf.SetFont("droid", "", 8)
		pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])
		for _, col := range cols {
			pdf.SetXY(x+padX, y0+1.4)
			pdf.CellFormat(col.width-2*padX, headerH-2.8, col.header, "", 0, "L", false, 0, "")
			x += col.width
		}
		pdf.SetY(y0 + headerH)
	}

	if len(rows) == 0 {
		pdf.SetFont("droid", "", 9)
		pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])
		pdf.SetX(pdfMarginX)
		pdf.CellFormat(tableW, 6, "（本期无记录）", "", 0, "L", false, 0, "")
		pdf.Ln(6)
		return
	}

	drawHeader()
	alt := false
	pdf.SetFont("droid", "", cellFont)
	for _, r := range rows {
		// 逐列拆行，记录每列的行数组与全行最大行数
		lines := make([][]string, len(r))
		maxLines := 1
		for i, cell := range r {
			ls := pdf.SplitText(cell, cols[i].width-2*padX)
			if len(ls) == 0 {
				ls = []string{""}
			}
			lines[i] = ls
			if len(ls) > maxLines {
				maxLines = len(ls)
			}
		}
		rowH := float64(maxLines)*lineH + 2*padY
		if pdf.GetY()+rowH > pageH-pdfBottom {
			pdf.AddPage()
			drawHeader()
			alt = false // 新页斑马纹从头开始
		}
		y0 := pdf.GetY()
		if alt {
			pdf.SetFillColor(pdfRowAlt[0], pdfRowAlt[1], pdfRowAlt[2])
			pdf.Rect(pdfMarginX, y0, tableW, rowH, "F")
		}
		alt = !alt
		pdf.SetTextColor(pdfInk[0], pdfInk[1], pdfInk[2])
		x := pdfMarginX
		for i := range r {
			for li, ln := range lines[i] {
				pdf.SetFont("droid", "", cellFont)
				pdf.SetXY(x+padX, y0+padY+float64(li)*lineH)
				pdf.CellFormat(cols[i].width-2*padX, lineH, ln, "", 0, "L", false, 0, "")
			}
			x += cols[i].width
		}
		pdf.SetDrawColor(pdfHair[0], pdfHair[1], pdfHair[2])
		pdf.SetLineWidth(0.15)
		pdf.Line(pdfMarginX, y0+rowH, pdfMarginX+tableW, y0+rowH)
		pdf.SetY(y0 + rowH)
	}
}

// drawBabyCard 档案信息卡（宝宝名 / 性别 / 出生日期 / 月龄 / 范围 / 导出人）
func drawBabyCard(pdf *fpdf.Fpdf, d *reportData, pageW, pageH float64) {
	ensureSpace(pdf, 34, pageH)
	w := pageW - 2*pdfMarginX
	y0 := pdf.GetY()
	pdf.SetFillColor(pdfTile[0], pdfTile[1], pdfTile[2])
	pdf.RoundedRect(pdfMarginX, y0, w, 30, 3, "1111", "F")
	ins := pdfMarginX + 6

	pdf.SetFont("droid", "", 8.5)
	pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])

	birth := "--"
	age := "--"
	if !d.BirthDate.IsZero() {
		birth = d.BirthDate.Format("2006-01-02")
		m := int(monthsBetween(d.BirthDate, d.Generated))
		age = fmt.Sprintf("%d个月 · 出生 %d天", m, calDayDiff(d.BirthDate, d.Generated))
	}
	rng := "全部记录"
	if d.StartStr != "" {
		rng = fmt.Sprintf("近 %d 天", d.Days)
	}
	owner := d.Owner
	if owner == "" {
		owner = "--"
	}
	tz := time.FixedZone("user", d.TzOffset*60)

	pdf.SetXY(ins, y0+6)
	pdf.CellFormat(w-12, 4.5, "出生日期 "+birth+" · 月龄 "+age, "", 0, "L", false, 0, "")
	pdf.SetXY(ins, y0+12.5)
	pdf.CellFormat(w-12, 4.5, "记录范围 "+rng+" · 导出人 "+owner, "", 0, "L", false, 0, "")
	pdf.SetXY(ins, y0+19)
	pdf.CellFormat(w-12, 4.5, "导出时间 "+d.Generated.Format("2006-01-02 15:04")+"（"+tz.String()+"）", "", 0, "L", false, 0, "")

	pdf.SetY(y0 + 32)
}

// drawOverviewTiles 概览统计格子（4×3）。
func drawOverviewTiles(pdf *fpdf.Fpdf, d *reportData, pageW, pageH float64) {
	ensureSpace(pdf, 3*18, pageH)
	w := pageW - 2*pdfMarginX
	cols, gap := 4, 4.0
	tw := (w - float64(cols-1)*gap) / float64(cols)
	tiles := overviewTiles(d)
	startY := pdf.GetY()
	rows := (len(tiles) + cols - 1) / cols
	for i, t := range tiles {
		x := pdfMarginX + float64(i%cols)*(tw+gap)
		y := startY + float64(i/cols)*18
		pdf.SetFillColor(pdfTile[0], pdfTile[1], pdfTile[2])
		pdf.RoundedRect(x, y, tw, 14, 2.5, "1111", "F")
		pdf.SetFont("droid", "", 6.8)
		pdf.SetTextColor(pdfSub[0], pdfSub[1], pdfSub[2])
		pdf.SetXY(x+2.5, y+1.2)
		pdf.CellFormat(tw-5, 3.6, t.label, "", 0, "L", false, 0, "")
		pdf.SetFont("droid", "", 11)
		pdf.SetTextColor(pdfInk[0], pdfInk[1], pdfInk[2])
		pdf.SetXY(x+2.5, y+5.4)
		pdf.CellFormat(tw-5, 7, t.value, "", 0, "L", false, 0, "")
	}
	pdf.SetY(startY + float64(rows)*18)
}

// drawGrowthCompare 成长对照：最新一次测量按 WS/T 423-2022 表1 五级评价 + 连续百分位。
func drawGrowthCompare(pdf *fpdf.Fpdf, d *reportData, pageW, pageH float64) {
	if d.BirthDate.IsZero() || len(d.Growth) == 0 {
		return
	}
	latest := d.Growth[0]
	gender := d.Gender
	if gender != "male" && gender != "female" {
		gender = "female"
	}
	month := monthsBetween(d.BirthDate, latest.T)
	if month < 0 {
		return
	}

	type item struct {
		label string
		value float64
		unit  string
		grade string
		pct   float64
		prec  int
	}
	var items []item
	for _, it := range []struct {
		label, metric string
		value         float64
		unit          string
		prec          int
	}{
		{"身高", "height", latest.H, "cm", 1},
		{"体重", "weight", latest.W, "kg", 2},
		{"头围", "head", latest.Hd, "cm", 1},
	} {
		if it.value <= 0 {
			continue
		}
		items = append(items, item{
			label: it.label,
			value: it.value,
			unit:  it.unit,
			grade: growthGrade(gender, d.BirthDate, latest.T, it.metric, it.value),
			pct:   growthPercentile(gender, d.BirthDate, latest.T, it.metric, it.value),
			prec:  it.prec,
		})
	}
	if len(items) == 0 {
		return
	}

	sub := fmt.Sprintf("最新测量生长对照 · 测量月龄 %s", ageMonthText(month))
	if d.Gender != "male" && d.Gender != "female" {
		sub += "（性别保密，暂按女宝标准）"
	}

	// 先把各指标文案量好，按可用宽度预排出行数，据此决定卡片高度；渲染时用
	// 显式 itemY/xi 落点，避免此前 subtitle 与指标共用同一 GetY() 造成的重叠。
	innerW := pageW - 2*pdfMarginX - 12
	type metricLine struct {
		it  item
		txt string
		w   float64
	}
	var metrics []metricLine
	rowCount := 1
	acc := 0.0
	for _, it := range items {
		pdf.SetFont("droid", "", 10)
		txt := fmt.Sprintf("%s %.*f%s", it.label, it.prec, it.value, it.unit)
		if it.grade == "" {
			txt += " · 无对照区间"
		} else if it.pct > 0 {
			txt += fmt.Sprintf(" · P%.0f · %s", it.pct, it.grade)
		} else {
			txt += " · " + it.grade
		}
		ml := metricLine{it: it, txt: txt, w: pdf.GetStringWidth(txt) + 4}
		metrics = append(metrics, ml)
		if acc > 0 && acc+ml.w > innerW {
			rowCount++
			acc = 0
		}
		acc += ml.w + 2
	}
	cardH := 12.0 + float64(rowCount)*6.5 + 2.0

	ensureSpace(pdf, cardH+2, pageH)
	w := pageW - 2*pdfMarginX
	y0 := pdf.GetY()
	pdf.SetFillColor(pdfTile[0], pdfTile[1], pdfTile[2])
	pdf.RoundedRect(pdfMarginX, y0, w, cardH, 3, "1111", "F")
	ins := pdfMarginX + 6

	pdf.SetFont("droid", "", 8)
	pdf.SetTextColor(pdfInk[0], pdfInk[1], pdfInk[2])
	pdf.SetXY(ins, y0+4)
	pdf.CellFormat(innerW, 4.5, sub, "", 0, "L", false, 0, "")

	xi := ins
	itemY := y0 + 12
	for _, ml := range metrics {
		if xi > ins && xi+ml.w > ins+innerW {
			xi = ins
			itemY += 6.5
		}
		if ml.it.grade != "" {
			pdf.SetFillColor(gradeColor(ml.it.grade)[0], gradeColor(ml.it.grade)[1], gradeColor(ml.it.grade)[2])
			pdf.Rect(xi, itemY+1.6, 2.2, 2.2, "F")
		}
		pdf.SetFont("droid", "", 10)
		pdf.SetTextColor(pdfInk[0], pdfInk[1], pdfInk[2])
		pdf.SetXY(xi+3, itemY)
		pdf.CellFormat(ml.w, 5, ml.txt, "", 0, "L", false, 0, "")
		xi += ml.w + 2
	}
	pdf.SetY(y0 + cardH + 2)
}

func gradeColor(g string) [3]int {
	switch g {
	case "上", "下":
		return pdfDanger
	case "中":
		return pdfSuccess
	}
	return pdfSub // 中上 / 中下：对称灰色
}

// ---- 行数据渲染 ----

func feedingRows(d *reportData, loc *time.Location) [][]string {
	var out [][]string
	for _, f := range d.Feeding {
		typ := map[string]string{"breast": "母乳", "bottle": "瓶喂", "formula": "配方奶"}[f.Type]
		if typ == "" {
			typ = f.Type
		}
		amt, dur := "", ""
		if f.Amt > 0 {
			amt = strconv.Itoa(f.Amt) + " ml"
		}
		if f.Dur > 0 {
			dur = strconv.Itoa(f.Dur) + " 分钟"
		}
		out = append(out, []string{f.T.In(loc).Format("2006-01-02 15:04"), typ, amt, dur, f.Note})
	}
	return out
}

func diaperRows(d *reportData, loc *time.Location) [][]string {
	var out [][]string
	for _, x := range d.Diaper {
		typ := map[string]string{"pee": "小便", "poop": "大便", "mixed": "混合"}[x.Type]
		if typ == "" {
			typ = x.Type
		}
		out = append(out, []string{x.T.In(loc).Format("2006-01-02 15:04"), typ, x.Note})
	}
	return out
}

func spanRows(rows []spanRow, loc *time.Location) [][]string {
	var out [][]string
	for _, r := range rows {
		end, dur := "—", spanOf(r)
		if !r.Ongoing {
			end = r.End.In(loc).Format("2006-01-02 15:04")
		}
		out = append(out, []string{r.Start.In(loc).Format("2006-01-02 15:04"), end, dur, r.Note})
	}
	return out
}

func tempRows(d *reportData, loc *time.Location) [][]string {
	var out [][]string
	for _, t := range d.Temp {
		locTxt := t.Loc
		if locTxt == "" {
			locTxt = "—"
		}
		out = append(out, []string{t.T.In(loc).Format("2006-01-02 15:04"), fmt.Sprintf("%.1f °C", t.Val), locTxt, t.Note})
	}
	return out
}

func suppRows(d *reportData, loc *time.Location) [][]string {
	var out [][]string
	for _, s := range d.Supplement {
		dose := "—"
		if s.Val > 0 {
			dose = fmt.Sprintf("%.1f %s", s.Val, s.Unit)
		}
		out = append(out, []string{s.T.In(loc).Format("2006-01-02 15:04"), s.Name, dose, s.Note})
	}
	return out
}

func growthRows(d *reportData) [][]string {
	var out [][]string
	for _, g := range d.Growth {
		age := "--"
		if !d.BirthDate.IsZero() {
			if m := monthsBetween(d.BirthDate, g.T); m >= 0 {
				age = ageMonthText(m)
			}
		}
		h, w, hd := "", "", ""
		if g.H > 0 {
			h = fmt.Sprintf("%.1f cm", g.H)
		}
		if g.W > 0 {
			w = fmt.Sprintf("%.2f kg", g.W)
		}
		if g.Hd > 0 {
			hd = fmt.Sprintf("%.1f cm", g.Hd)
		}
		out = append(out, []string{g.T.Format("2006-01-02"), age, h, w, hd, g.Note})
	}
	return out
}

// ---- 概览统计 ----

type tileSpec struct {
	label, value string
}

func overviewTiles(d *reportData) []tileSpec {
	var tiles []tileSpec

	fc := len(d.Feeding)
	tiles = append(tiles, tileSpec{"喂奶次数", strconv.Itoa(fc) + " 次"})
	totalML := 0
	var feedTimes []time.Time
	for _, f := range d.Feeding {
		totalML += f.Amt
		feedTimes = append(feedTimes, f.T)
	}
	tiles = append(tiles, tileSpec{"总奶量", strconv.Itoa(totalML) + " ml"})
	tiles = append(tiles, tileSpec{"平均间隔", gapText(feedTimes)})
	tiles = append(tiles, tileSpec{"尿布次数", strconv.Itoa(len(d.Diaper)) + " 次"})

	sCnt, sTot, sAvg := spanStats(d.Sleep, d.Generated)
	tiles = append(tiles, tileSpec{"睡眠时长", durText(sTot)})
	tiles = append(tiles, tileSpec{"睡眠次数", strconv.Itoa(sCnt) + " 次"})
	tiles = append(tiles, tileSpec{"平均单次", durText(sAvg)})

	oCnt, oTot, oAvg := spanStats(d.Outdoor, d.Generated)
	tiles = append(tiles, tileSpec{"户外时长", durText(oTot)})
	tiles = append(tiles, tileSpec{"户外次数", strconv.Itoa(oCnt) + " 次"})
	tiles = append(tiles, tileSpec{"户外平均", durText(oAvg)})

	var tMax float64
	tFever := 0
	for _, t := range d.Temp {
		if t.Val > tMax {
			tMax = t.Val
		}
		if t.Val >= 37.5 {
			tFever++
		}
	}
	tiles = append(tiles, tileSpec{"测温次数", strconv.Itoa(len(d.Temp)) + " 次"})
	tiles = append(tiles, tileSpec{"发烧记录", strconv.Itoa(tFever) + " 次"})

	var suppTimes []time.Time
	names := map[string]bool{}
	for _, s := range d.Supplement {
		suppTimes = append(suppTimes, s.T)
		if s.Name != "" {
			names[s.Name] = true
		}
	}
	tiles = append(tiles, tileSpec{"补剂次数", strconv.Itoa(len(d.Supplement)) + " 次"})
	tiles = append(tiles, tileSpec{"补剂种类", strconv.Itoa(len(names)) + " 种"})

	return tiles
}

func spanStats(rows []spanRow, now time.Time) (count int, total, avg float64) {
	count = len(rows)
	for _, r := range rows {
		end := r.End
		if r.Ongoing || end.IsZero() {
			end = now
		}
		total += end.Sub(r.Start).Minutes()
	}
	if count > 0 {
		avg = total / float64(count)
	}
	return
}

func gapText(times []time.Time) string {
	if len(times) < 2 {
		return "--"
	}
	t := make([]time.Time, len(times))
	copy(t, times)
	sort.Slice(t, func(i, j int) bool { return t[i].Before(t[j]) })
	var total float64
	for i := 1; i < len(t); i++ {
		total += t[i].Sub(t[i-1]).Minutes()
	}
	return durText(total / float64(len(t)-1))
}

func durText(min float64) string {
	if min < 1 {
		return "--"
	}
	if min < 60 {
		return fmt.Sprintf("%.0f分钟", min)
	}
	return fmt.Sprintf("%.1f小时", min/60)
}

// ---- 文案辅助 ----

func genderText(g string) string {
	switch g {
	case "male":
		return "男宝"
	case "female":
		return "女宝"
	}
	return "保密"
}

func calDayDiff(a, b time.Time) int {
	y, m, d := a.Date()
	ay, am, ad := b.Date()
	t1 := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(ay, am, ad, 0, 0, 0, 0, time.UTC)
	return int(math.Floor(t2.Sub(t1).Hours() / 24))
}

func ageMonthText(m float64) string {
	if m < 0 {
		return "--"
	}
	return strconv.Itoa(int(m)) + "个月"
}