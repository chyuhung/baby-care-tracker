package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestRateLimit 验证固定窗口限流：窗口内第 limit 次之后返回 429，窗口重置后恢复。
func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ping", RateLimit(2, time.Minute), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// 同 IP 连打 3 次：前 2 次放行，第 3 次 429
	for i := 1; i <= 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "203.0.113.7:5678"
		r.ServeHTTP(rec, req)

		if i <= 2 {
			if rec.Code != http.StatusOK {
				t.Fatalf("第 %d 次请求（窗口内）应放行 200，实际 %d", i, rec.Code)
			}
		} else if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("第 %d 次请求（超限）应 429，实际 %d", i, rec.Code)
		}
	}

	// 不同 IP 不受影响：独立配额
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "198.51.100.9:5678"
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("不同 IP 应独立计数并放行 200，实际 %d", rec.Code)
	}
}

// TestRateLimitWindowReset 验证窗口过期后计数重置（旧窗口的计数不残留）。
func TestRateLimitWindowReset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ping", RateLimit(1, 50*time.Millisecond), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	send := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "203.0.113.7:5678"
		r.ServeHTTP(rec, req)
		return rec
	}

	// 窗口 50ms，limit 1：第一次放行，紧接着第二次 429
	if send().Code != http.StatusOK {
		t.Fatal("窗口内首次请求应放行")
	}
	if send().Code != http.StatusTooManyRequests {
		t.Fatal("窗口内第二次请求应 429")
	}

	// 等窗口过期，恢复放行
	time.Sleep(80 * time.Millisecond)
	if send().Code != http.StatusOK {
		t.Fatal("窗口重置后应恢复放行")
	}
}

// TestGuardCSVCell 验证公式注入防护只作用于首字符危险单元格。
func TestGuardCSVCell(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"正常备注":            "正常备注",
		"=SUM(A1:A2)":   "'=SUM(A1:A2)",
		"+cmd|calc":     "'+cmd|calc",
		"-1+1":          "'-1+1",
		"@SUM":          "'@SUM",
		"\t开头":          "'\t开头",
		"已经=引号保护的":       "已经=引号保护的", // 仅首字符触发，等号在中间不动
		"abc=1":         "abc=1",
	}
	for in, want := range cases {
		if got := guardCSVCell(in); got != want {
			t.Errorf("guardCSVCell(%q) = %q，期望 %q", in, got, want)
		}
	}
}