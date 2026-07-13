package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

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
