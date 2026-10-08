package handlers

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBuildPDF_Empty(t *testing.T) {
	d := &reportData{
		BabyName: "测试宝宝",
		Gender:   "female",
		Generated: time.Now().In(time.FixedZone("user", 8*60)),
	}
	out, err := buildPDF(d)
	if err != nil {
		t.Fatalf("buildPDF empty: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF")) {
		t.Fatalf("output is not a PDF (header %q)", out[:min(8, len(out))])
	}
	if len(out) < 1000 {
		t.Fatalf("PDF too small: %d bytes", len(out))
	}
}

func TestBuildPDF_Rich(t *testing.T) {
	loc := time.FixedZone("user", 8*60)
	now := time.Now().In(loc)
	d := &reportData{
		BabyName:  "龙宝",
		Gender:    "male",
		BirthDate: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		Owner:     "mama",
		TzOffset:  8 * 60,
		Generated: now,
		Feeding: []feedingRow{
			{T: now.Add(-time.Hour), Type: "bottle", Amt: 120, Dur: 15, Note: "无"},
			{T: now.Add(-4 * time.Hour), Type: "breast", Side: "both"},
		},
		Diaper: []diaperRow{
			{T: now.Add(-30 * time.Minute), Type: "mixed", Note: "颜色正常"},
		},
		Sleep: []spanRow{
			{Start: now.Add(-3 * time.Hour), End: now.Add(-2 * time.Hour), Note: ""},
			{Start: now.Add(-time.Hour), Ongoing: true, Note: "正在睡"},
		},
		Temp: []tempRow{
			{T: now.Add(-2 * time.Hour), Val: 37.8, Loc: "腋下", Note: "偏高"},
		},
		Outdoor: []spanRow{
			{Start: now.Add(-5 * time.Hour), End: now.Add(-4 * time.Hour)},
		},
		Supplement: []suppRow{
			{T: now.Add(-6 * time.Hour), Name: "维生素D", Val: 400, Unit: "IU"},
		},
		Growth: []growthRow{
			{T: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC), H: 52, W: 4.25, Hd: 36.1},
		},
	}
	out, err := buildPDF(d)
	if err != nil {
		t.Fatalf("buildPDF rich: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF")) {
		t.Fatalf("output is not a PDF")
	}
	if len(out) < 5000 {
		t.Fatalf("rich PDF too small: %d bytes", len(out))
	}
	// 冒烟：多页也不崩溃（模拟大量记录）
	big := *d
	for i := 0; i < 300; i++ {
		big.Feeding = append(big.Feeding, feedingRow{T: now.Add(-time.Duration(i) * time.Minute), Type: "bottle", Amt: 90})
	}
	if _, err := buildPDF(&big); err != nil {
		t.Fatalf("buildPDF multi-page: %v", err)
	}
}

func TestCSVRows_SortedAndDetail(t *testing.T) {
	loc := time.FixedZone("user", 8*60)
	now := time.Now().In(loc)
	d := &reportData{
		Feeding: []feedingRow{
			{T: now, Type: "breast", Side: "left", Note: "备注"},
			{T: now.Add(-time.Hour), Type: "bottle", Amt: 120, Dur: 10},
		},
		Sleep: []spanRow{
			{Start: now.Add(-2 * time.Hour), End: now.Add(-1 * time.Hour)},
			{Start: now.Add(-30 * time.Minute), Ongoing: true},
		},
		Growth: []growthRow{
			{T: now.Add(-24 * time.Hour), H: 52, W: 4.2, Hd: 36},
		},
	}
	rows := d.csvRows()
	if len(rows) != 5 {
		t.Fatalf("want 5 rows, got %d", len(rows))
	}
	for i := 1; i < len(rows); i++ {
		if !rows[i-1].t.After(rows[i].t) {
			t.Fatalf("rows not sorted descending at %d", i)
		}
	}
	got := map[string][]string{}
	for _, r := range rows {
		got[r.kind] = append(got[r.kind], r.detail)
	}
	if len(got["喂奶"]) != 2 || got["喂奶"][0] != "母乳 (左)" {
		t.Errorf("feeding details got %v", got["喂奶"])
	}
	if got["喂奶"][1] != "瓶喂 120ml 10分钟" {
		t.Errorf("bottle detail got %q", got["喂奶"][1])
	}
	if len(got["睡眠"]) != 2 || got["睡眠"][0] != "进行中" || got["睡眠"][1] != "60分钟" {
		t.Errorf("sleep details got %v", got["睡眠"])
	}
	if got["成长"][0] != "身高 52.0cm 体重 4.20kg 头围 36.0cm" {
		t.Errorf("growth detail got %q", got["成长"][0])
	}
	// 进行中睡眠单测：spanOf
	if spanOf(d.Sleep[1]) != "进行中" {
		t.Errorf("ongoing span got %q", spanOf(d.Sleep[1]))
	}
	// 长备注跨页多行必须能进 PDF
	long := *d
	long.Feeding[0].Note = strings.Repeat("这是很长的备注说明", 40)
	if _, err := buildPDF(&long); err != nil {
		t.Fatalf("buildPDF long note: %v", err)
	}
}