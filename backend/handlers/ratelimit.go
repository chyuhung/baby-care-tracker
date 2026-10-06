package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateBucket 单个 IP 在某一窗口内的计数。
type rateBucket struct {
	count       int
	windowStart time.Time
}

// RateLimit 返回一个每 IP 固定窗口限流中间件：同一窗口内超过 limit 次请求返回 429。
// 每个挂载点（RateLimit 调用）持有独立的状态，因此可按路由粒度设置不同配额。
// 无外部依赖，进程内存储；单机部署（本项目形态）下足够，多实例需换共享存储。
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*rateBucket)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b, ok := buckets[ip]
		if !ok {
			b = &rateBucket{windowStart: now}
			buckets[ip] = b
		} else if now.Sub(b.windowStart) >= window {
			// 窗口已过，重置计数
			b.windowStart = now
			b.count = 0
		}
		b.count++
		tooMany := b.count > limit

		// 惰性清理：只防极端情况下 map 无限膨胀（正常家庭场景 IP 极少）
		if len(buckets) > 2048 {
			for k, bb := range buckets {
				if now.Sub(bb.windowStart) >= window {
					delete(buckets, k)
				}
			}
		}
		mu.Unlock()

		if tooMany {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			return
		}
		c.Next()
	}
}