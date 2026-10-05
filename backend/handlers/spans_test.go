package handlers

import (
	"baby-care-tracker/database"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// 用内存库替换包级 DB，跑 sumSpansByDay 的端到端行为（含 SQL 重叠谓词）。
func withTestDB(t *testing.T, ddl []string) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	for _, s := range ddl {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("ddl: %v", err)
		}
	}
	orig := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = orig
		db.Close()
	})
}

const spanDDL = `
CREATE TABLE sleep_records (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	baby_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	started_at DATETIME NOT NULL,
	ended_at DATETIME,
	note TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE outdoor_records (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	baby_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	started_at DATETIME NOT NULL,
	ended_at DATETIME,
	note TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

func insSpan(t *testing.T, table string, babyID int64, started, ended string) {
	t.Helper()
	var v interface{}
	if ended != "" {
		v = ended
	}
	if _, err := database.DB.Exec(
		"INSERT INTO "+table+" (baby_id, user_id, started_at, ended_at) VALUES (?, 1, ?, ?)",
		babyID, started, v,
	); err != nil {
		t.Fatalf("insert %s: %v", table, err)
	}
}

func rfc(t *testing.T, loc *time.Location, s string) string {
	t.Helper()
	v, err := time.ParseInLocation(time.RFC3339, s, loc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return v.UTC().Format(time.RFC3339)
}

// 窗口重叠谓词 + 按日切分的端到端验证
func TestSumSpansByDay(t *testing.T) {
	withTestDB(t, []string{spanDDL})

	tz := time.FixedZone("user", 8*3600)
	loc := tz
	windowStart := rfc(t, loc, "2026-03-10T00:00:00+08:00")
	windowEnd := rfc(t, loc, "2026-03-12T00:00:00+08:00")

	cases := []struct {
		name    string
		started string
		ended   string
		expect  map[string]int
	}{
		{
			name:    "窗口内完全包含的同日记录",
			started: "2026-03-10T02:00:00+08:00",
			ended:   "2026-03-10T06:00:00+08:00",
			expect:  map[string]int{"2026-03-10": 240},
		},
		{
			name:    "窗口内跨夜：两侧都计入",
			started: "2026-03-10T20:00:00+08:00",
			ended:   "2026-03-11T06:00:00+08:00",
			expect:  map[string]int{"2026-03-10": 240, "2026-03-11": 360},
		},
		{
			// 核心回归：旧实现「started_at >= windowStart」会把这条整条丢掉，
			// 导致 03-11 统计为 0。
			// 返回值未按窗口裁剪，故 03-09 的键也存在——调用方只按自己的日期列表取值。
			name:    "窗口前开始、窗口内结束 —— 旧实现会漏掉",
			started: "2026-03-09T20:00:00+08:00",
			ended:   "2026-03-10T06:00:00+08:00",
			expect:  map[string]int{"2026-03-09": 240, "2026-03-10": 360},
		},
		{
			name:    "窗口内开始、窗口后结束 —— 只计窗口内部分",
			started: "2026-03-11T22:00:00+08:00",
			ended:   "2026-03-12T02:00:00+08:00",
			expect:  map[string]int{"2026-03-11": 120, "2026-03-12": 120},
		},
		{
			name:    "完全在窗口前 —— 不计入",
			started: "2026-03-08T20:00:00+08:00",
			ended:   "2026-03-09T06:00:00+08:00",
			expect:  nil,
		},
		{
			name:    "完全在窗口后 —— 不计入",
			started: "2026-03-13T02:00:00+08:00",
			ended:   "2026-03-13T06:00:00+08:00",
			expect:  nil,
		},
		{
			// 脏数据容错：end <= start 应被丢弃而非产生负数
			name:    "结束早于开始（脏数据）—— 不计入",
			started: "2026-03-10T10:00:00+08:00",
			ended:   "2026-03-10T09:00:00+08:00",
			expect:  nil,
		},
	}

	// 每个用例用一个独立的 babyID，避免用例间互相污染累计值
	for i, c := range cases {
		babyID := int64(i + 1)
		insSpan(t, "sleep_records", babyID, rfc(t, loc, c.started), rfc(t, loc, c.ended))
		// 每条同步写入户外，用于验证两表行为一致（sumSpansByDay 是共用实现）
		insSpan(t, "outdoor_records", babyID, rfc(t, loc, c.started), rfc(t, loc, c.ended))

		for _, table := range []string{"sleep_records", "outdoor_records"} {
			got, err := sumSpansByDay(table, babyID, windowStart, windowEnd, loc)
			if err != nil {
				t.Fatalf("case %d %s: %v", i, table, err)
			}
			if len(got) != len(c.expect) {
				t.Fatalf("case %d %s: len got %v want %v", i, table, got, c.expect)
			}
			for k, v := range c.expect {
				if got[k] != v {
					t.Fatalf("case %d %s: %s got %d want %d (full=%v)", i, table, k, got[k], v, got)
				}
			}
		}
	}
}

// 进行中的记录（ended_at 为 NULL）以 now 收尾，必须被计入
func TestSumSpansByDay_Ongoing(t *testing.T) {
	withTestDB(t, []string{spanDDL})
	tz := time.FixedZone("user", 8*3600)
	const babyID = 1

	// 3 小时前开始、至今未结束
	started := time.Now().Add(-3 * time.Hour)
	windowStart := rfc(t, tz, started.Add(-24*time.Hour).Format("2006-01-02T15:04:05Z07:00"))
	windowEnd := rfc(t, tz, time.Now().Add(24*time.Hour).Format("2006-01-02T15:04:05Z07:00"))

	insSpan(t, "sleep_records", babyID, started.UTC().Format(time.RFC3339), "")

	got, err := sumSpansByDay("sleep_records", babyID, windowStart, windowEnd, tz)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	total := 0
	for _, v := range got {
		total += v
	}
	// 约 180 分钟，允许 ±2 分钟执行耗时
	if total < 178 || total > 182 {
		t.Fatalf("ongoing 应累计约 180 分钟，实得 %d (%v)", total, got)
	}
}

func TestSumSpansByDay_RejectsBadTable(t *testing.T) {
	withTestDB(t, []string{spanDDL})
	if _, err := sumSpansByDay("feeding_records; DROP TABLE users--", 1, "", "", time.UTC); err == nil {
		t.Fatal("非法表名必须报错")
	}
}

// 多条记录落在同一天必须累加而非覆盖
func TestSumSpansByDay_Accumulates(t *testing.T) {
	withTestDB(t, []string{spanDDL})
	tz := time.FixedZone("user", 8*3600)
	const babyID = 1
	insSpan(t, "sleep_records", babyID, rfc(t, tz, "2026-03-10T01:00:00+08:00"), rfc(t, tz, "2026-03-10T02:00:00+08:00"))
	insSpan(t, "sleep_records", babyID, rfc(t, tz, "2026-03-10T03:00:00+08:00"), rfc(t, tz, "2026-03-10T04:30:00+08:00"))
	insSpan(t, "sleep_records", babyID, rfc(t, tz, "2026-03-10T23:00:00+08:00"), rfc(t, tz, "2026-03-11T01:00:00+08:00"))

	got, err := sumSpansByDay("sleep_records", babyID,
		rfc(t, tz, "2026-03-10T00:00:00+08:00"), rfc(t, tz, "2026-03-12T00:00:00+08:00"), tz)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["2026-03-10"] != 60+90+60 {
		t.Fatalf("03-10 应为 210，实得 %d (%v)", got["2026-03-10"], got)
	}
	if got["2026-03-11"] != 60 {
		t.Fatalf("03-11 应为 60，实得 %d (%v)", got["2026-03-11"], got)
	}
}

// 其他宝宝的记录不得被计入
func TestSumSpansByDay_BabyIsolation(t *testing.T) {
	withTestDB(t, []string{spanDDL})
	tz := time.FixedZone("user", 8*3600)
	insSpan(t, "sleep_records", 1, rfc(t, tz, "2026-03-10T01:00:00+08:00"), rfc(t, tz, "2026-03-10T02:00:00+08:00"))
	insSpan(t, "sleep_records", 2, rfc(t, tz, "2026-03-10T01:00:00+08:00"), rfc(t, tz, "2026-03-10T09:00:00+08:00"))

	got, err := sumSpansByDay("sleep_records", 1,
		rfc(t, tz, "2026-03-10T00:00:00+08:00"), rfc(t, tz, "2026-03-12T00:00:00+08:00"), tz)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["2026-03-10"] != 60 {
		t.Fatalf("应只计自己宝宝的 60 分钟，实得 %d (%v)", got["2026-03-10"], got)
	}
}
