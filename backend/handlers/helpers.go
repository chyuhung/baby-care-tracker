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
