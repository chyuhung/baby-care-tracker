package handlers

import (
	"sort"
	"testing"
	"time"
)

// helper：把 "2026-01-01 20:00:00" 按 loc 解析为时刻
func at(t *testing.T, loc *time.Location, s string) time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

func TestSplitSpanByLocalDay(t *testing.T) {
	tz := time.FixedZone("test", 8*3600)

	cases := []struct {
		name   string
		start  string
		end    string
		expect map[string]int
	}{
		{
			name:   "同日",
			start:  "2026-03-10 02:00:00",
			end:    "2026-03-10 06:00:00",
			expect: map[string]int{"2026-03-10": 240},
		},
		{
			// 20:00 -> 次日06:00 共 10h = 600min，拆成 240 + 360
			name:   "跨天常规 20:00 -> 次日06:00 = 240+360",
			start:  "2026-03-10 20:00:00",
			end:    "2026-03-11 06:00:00",
			expect: map[string]int{"2026-03-10": 240, "2026-03-11": 360},
		},
		{
			name:   "跨天常规 18:00 -> 次日06:00 = 360+360",
			start:  "2026-03-10 18:00:00",
			end:    "2026-03-11 06:00:00",
			expect: map[string]int{"2026-03-10": 360, "2026-03-11": 360},
		},
		{
			name:   "起点恰为0点 → 整段归该日",
			start:  "2026-03-10 00:00:00",
			end:    "2026-03-11 00:00:00",
			expect: map[string]int{"2026-03-10": 1440},
		},
		{
			name:   "终点恰为次日0点 → 不产生次日0分钟条目",
			start:  "2026-03-10 20:00:00",
			end:    "2026-03-11 00:00:00",
			expect: map[string]int{"2026-03-10": 240},
		},
		{
			// 累计取整差分的核心用例：
			//   20:00:40 -> 次日00:00:00 = 239.33min -> 239
			//   次日00:00:00 -> 00:20:20   =  20.33min -> 差分 259-239 = 20
			// 逐段独立 int() 截断会得到 239+0=239，丢 20 分钟。
			name:   "末段不足/超过1分钟仍守恒（累计差分）",
			start:  "2026-03-10 20:00:40",
			end:    "2026-03-11 00:20:20",
			expect: map[string]int{"2026-03-10": 239, "2026-03-11": 20},
		},
		{
			name:   "跨3天（忘记结束）",
			start:  "2026-03-10 22:00:00",
			end:    "2026-03-13 06:00:00",
			expect: map[string]int{"2026-03-10": 120, "2026-03-11": 1440, "2026-03-12": 1440, "2026-03-13": 360},
		},
		{
			name:   "跨月",
			start:  "2026-03-30 23:00:00",
			end:    "2026-04-01 01:00:00",
			expect: map[string]int{"2026-03-30": 60, "2026-03-31": 1440, "2026-04-01": 60},
		},
		{
			name:   "亚分钟总时长 <1min → 空（不产生0条目）",
			start:  "2026-03-10 10:00:00",
			end:    "2026-03-10 10:00:30",
			expect: map[string]int{},
		},
		{
			name:   "终点等于起点 → 空",
			start:  "2026-03-10 10:00:00",
			end:    "2026-03-10 10:00:00",
			expect: map[string]int{},
		},
		{
			name:   "终点早于起点（脏数据）→ 空",
			start:  "2026-03-10 10:00:00",
			end:    "2026-03-10 09:00:00",
			expect: map[string]int{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := splitSpanByLocalDay(at(t, tz, c.start), at(t, tz, c.end), tz)
			if len(got) != len(c.expect) {
				t.Fatalf("len mismatch: got %v want %v", got, c.expect)
			}
			for k, v := range c.expect {
				if got[k] != v {
					t.Fatalf("date %s: got %d want %d (full=%v)", k, got[k], v, got)
				}
			}
		})
	}
}

func TestSplitSpanByLocalDay_ZeroTime(t *testing.T) {
	tz := time.FixedZone("test", 8*3600)
	if got := splitSpanByLocalDay(time.Time{}, at(t, tz, "2026-03-10 06:00:00"), tz); len(got) != 0 {
		t.Fatalf("zero start should give empty, got %v", got)
	}
	if got := splitSpanByLocalDay(at(t, tz, "2026-03-10 02:00:00"), time.Time{}, tz); len(got) != 0 {
		t.Fatalf("zero end should give empty, got %v", got)
	}
	if got := splitSpanByLocalDay(at(t, tz, "2026-03-10 02:00:00"), at(t, tz, "2026-03-10 06:00:00"), nil); len(got) != 0 {
		t.Fatalf("nil loc should give empty, got %v", got)
	}
}

// 性质测试：切分结果必须与「时区无关」且守恒。
func TestSplitSpanByLocalDay_Properties(t *testing.T) {
	offsets := []int{-14 * 3600, -8 * 3600, 0, 5*3600 + 1800, 8 * 3600, 14 * 3600}
	shapes := [][2]string{
		{"2026-03-10 02:00:00", "2026-03-10 02:00:01"},
		{"2026-03-10 02:00:00", "2026-03-10 23:59:59"},
		{"2026-03-10 23:59:59", "2026-03-11 00:00:00"},
		{"2026-03-10 20:00:40", "2026-03-11 00:20:20"},
		{"2026-03-10 00:00:00", "2026-03-13 00:00:00"},
		{"2026-02-28 23:00:00", "2026-03-01 01:00:00"}, // 2026 非闰年
		{"2026-03-31 23:30:00", "2026-04-01 00:30:00"},
	}

	for _, sh := range shapes {
		// 同一组墙钟时间在不同偏移下，日切结果应完全一致
		var ref map[string]int
		for _, off := range offsets {
			loc := time.FixedZone("t", off)
			st, _ := time.ParseInLocation("2006-01-02 15:04:05", sh[0], loc)
			et, _ := time.ParseInLocation("2006-01-02 15:04:05", sh[1], loc)
			got := splitSpanByLocalDay(st, et, loc)

			// 守恒：sum ≡ floor(总分钟)
			total := int(et.Sub(st).Minutes())
			sum := 0
			for _, v := range got {
				sum += v
			}
			if sum != total {
				t.Fatalf("%v offset=%d: sum=%d want floor(total)=%d (got %v)", sh, off, sum, total, got)
			}
			// 单日不可能超过 1440 分钟
			for k, v := range got {
				if v <= 0 || v > 1440 {
					t.Fatalf("%v offset=%d: %s=%d out of range (got %v)", sh, off, k, v, got)
				}
			}
			// 日期键必须连续无缺口
			keys := make([]string, 0, len(got))
			for k := range got {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for i := 1; i < len(keys); i++ {
				p, _ := time.ParseInLocation("2006-01-02", keys[i-1], loc)
				n, _ := time.ParseInLocation("2006-01-02", keys[i], loc)
				if n.Sub(p) != 24*time.Hour {
					t.Fatalf("%v offset=%d: 日期不连续 %s -> %s", sh, off, keys[i-1], keys[i])
				}
			}
			if ref == nil {
				ref = got
			} else if len(ref) != len(got) {
				t.Fatalf("%v offset=%d: 时区导致分段数不同 ref=%v got=%v", sh, off, ref, got)
			} else {
				for k, v := range ref {
					if got[k] != v {
						t.Fatalf("%v offset=%d: 时区导致日切不同 %s ref=%d got=%d", sh, off, k, v, got[k])
					}
				}
			}
		}
	}
}

// windowRangeUTC 必须与 lastNDates 的日期列表严格对齐
func TestWindowRangeUTC_AlignsWithLastNDates(t *testing.T) {
	for _, off := range []int{-8 * 3600, 0, 8 * 3600} {
		for _, days := range []int{7, 30} {
			start, end := windowRangeUTC(off, days)
			first := lastNDates(off, days)[0]
			last := lastNDates(off, days)[days-1]

			st, err := time.Parse(time.RFC3339, start)
			if err != nil {
				t.Fatalf("start not RFC3339: %v", err)
			}
			et, err := time.Parse(time.RFC3339, end)
			if err != nil {
				t.Fatalf("end not RFC3339: %v", err)
			}
			loc := time.FixedZone("user", off*60)

			// 窗口起点应等于首个日期的当地 0 点
			want, _ := time.ParseInLocation("2006-01-02", first, loc)
			if !st.In(loc).Equal(want) {
				t.Fatalf("off=%d days=%d: start %s != %s 00:00", off, days, st.In(loc), first)
			}
			// 窗口终点应等于末个日期次日的当地 0 点
			lastT, _ := time.ParseInLocation("2006-01-02", last, loc)
			wantEnd := lastT.AddDate(0, 0, 1)
			if !et.In(loc).Equal(wantEnd) {
				t.Fatalf("off=%d days=%d: end %s != %s", off, days, et.In(loc), wantEnd)
			}
			// 窗口总长必须恰为 days 个整日
			if days*24 != int(et.Sub(st).Hours()) {
				t.Fatalf("off=%d days=%d: window span %v not %d days", off, days, et.Sub(st), days)
			}
		}
	}
}
