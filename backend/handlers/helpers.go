package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// parseTime 解析各种 ISO 格式的时间字符串
func parseTime(s string) time.Time {
	// 有时区信息（含 Z 或 ±HH:MM）→ 直接解析为 UTC
	if strings.ContainsAny(s, "Zz") || len(s) > 19 {
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			t, err := time.Parse(layout, s)
			if err == nil {
				return t
			}
		}
	}
	// 无时区信息（旧数据，存储为 server local time）→ 按 local 解析再转 UTC
	t, err := time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
	if err == nil {
		return t.UTC()
	}
	return time.Time{}
}

// normalizeBirthDate 把出生日期规范为纯日历日（YYYY-MM-DD）。
// 出生日期是日历日而非时刻：存本地日期才能在跨时区、跨零点后保持同一天；
// 旧数据（RFC3339 UTC）按客户端时区换算回日历日。
func normalizeBirthDate(s string, tzOffset int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if len(s) >= 10 {
		if _, err := time.Parse("2006-01-02", s[:10]); err == nil && len(s) <= 10 {
			return s[:10]
		}
	}
	t := parseTime(s)
	if t.IsZero() {
		return s
	}
	return t.In(time.FixedZone("user", tzOffset*60)).Format("2006-01-02")
}

// localDateKey 返回时刻在 loc 时区下的日历日键（YYYY-MM-DD）
func localDateKey(t time.Time, loc *time.Location) string {
	l := t.In(loc)
	return fmt.Sprintf("%d-%02d-%02d", l.Year(), l.Month(), l.Day())
}

// localMidnight 返回时刻在 loc 时区下所属自然日的当地 0 点
func localMidnight(t time.Time, loc *time.Location) time.Time {
	l := t.In(loc)
	return time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, loc)
}

// splitSpanByLocalDay 把区间 [start, end) 按用户时区的自然日（0 点）切成
// map[日历日]分钟数。跨天的睡眠/户外必须用它，不能把整段时长记到起始日 ——
// 那样「前一天傍晚睡到次日凌晨」会全部计入前一天，次日统计为 0。
//
// 采用**累计取整差分**：每段 = int(段末 - start) - int(段初 - start)。
// 若逐段独立 int() 取整，末段不足 1 分钟会归 0，各段之和 < floor(总分钟)；
// 差分法保证 sum(各段) ≡ floor(总分钟)，且不会因四舍五入而溢出。
//
// 边界：
//   - end <= start 或任一为零值 → 空 map
//   - start 恰为 0 点 → 整段归该日
//   - end 恰为次日 0 点 → 不产生次日 0 分钟条目
//   - 自然日推进用 AddDate(0,0,1) 而非 Add(24h)：语义上正确，且在含 DST 的
//     Location 下不会漂移（getTzOffset 返回固定偏移，但不应依赖这一点）
func splitSpanByLocalDay(start, end time.Time, loc *time.Location) map[string]int {
	out := make(map[string]int)
	if loc == nil || start.IsZero() || end.IsZero() || !end.After(start) {
		return out
	}
	prevCum := 0
	for cursor := localMidnight(start, loc); ; cursor = cursor.AddDate(0, 0, 1) {
		segEnd := end
		if next := cursor.AddDate(0, 0, 1); next.Before(end) {
			segEnd = next
		}
		cum := int(segEnd.Sub(start).Minutes())
		if seg := cum - prevCum; seg > 0 {
			out[localDateKey(cursor, loc)] += seg
		}
		prevCum = cum
		if !segEnd.Before(end) {
			return out
		}
	}
}

// getTzOffset 从请求头中获取客户端时区偏移（分钟），默认0（UTC）
func getTzOffset(c *gin.Context) int {
	header := c.GetHeader("X-Timezone-Offset")
	if header == "" {
		return 0
	}
	offset, err := strconv.Atoi(header)
	if err != nil {
		return 0
	}
	return offset
}

// todayDateRange 返回今天在用户时区下的 UTC 起止时间字符串
func todayDateRange(tzOffset int) (start, end string) {
	loc := time.FixedZone("user", tzOffset*60)
	now := time.Now().In(loc)
	y, m, d := now.Date()
	startLocal := time.Date(y, m, d, 0, 0, 0, 0, loc)
	endLocal := startLocal.Add(24 * time.Hour)
	start = startLocal.UTC().Format(time.RFC3339)
	end = endLocal.UTC().Format(time.RFC3339)
	return
}

// lastNDates 返回最近 N 天在用户时区下的日期字符串列表（不含今天约整）
func lastNDates(tzOffset, n int) []string {
	loc := time.FixedZone("user", tzOffset*60)
	now := time.Now().In(loc)
	var dates []string
	for i := n - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		dates = append(dates, fmt.Sprintf("%d-%02d-%02d", d.Year(), d.Month(), d.Day()))
	}
	return dates
}

// daysAgoUTC 返回 N 天前 0 点在用户时区下的 UTC 起始时间
func daysAgoUTC(tzOffset, days int) string {
	loc := time.FixedZone("user", tzOffset*60)
	now := time.Now().In(loc)
	startLocal := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	startLocal = startLocal.AddDate(0, 0, -days+1)
	return startLocal.UTC().Format(time.RFC3339)
}

// windowRangeUTC 返回最近 N 天（含今天）在用户时区下的 UTC 起止边界 [start, end)，
// 与 lastNDates(tzOffset, N) 的日期列表严格对齐。
// 跨天区间记录（睡眠/户外）必须用这对边界做「重叠」筛选，不能只比 started_at：
// 昨天 20:00 开始、今天 06:00 结束的睡眠，其今天部分落在窗口内，
// 用 started_at >= start 会把整条记录排除，今天就统计为 0。
func windowRangeUTC(tzOffset, days int) (start, end string) {
	loc := time.FixedZone("user", tzOffset*60)
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	windowStart := todayStart.AddDate(0, 0, -days+1)
	return windowStart.UTC().Format(time.RFC3339), todayStart.AddDate(0, 0, 1).UTC().Format(time.RFC3339)
}

// spanOverlapFilter 返回「与 [start, end) 有重叠」的 SQL 条件（含占位符），
// 供区间型记录（睡眠/户外）复用；配合 args 顺序追加 start、end。
// 进行中（ended_at IS NULL）的记录视为延伸到无穷远，因此只要 started_at < end 即算重叠。
func spanOverlapFilter() string {
	return " AND started_at < ? AND (ended_at IS NULL OR ended_at > ?)"
}
