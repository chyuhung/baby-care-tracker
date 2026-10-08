package handlers

import (
	"baby-care-tracker/database"
	"time"
)

// ---- 导出的共享取数层（CSV 与 PDF 共用同一份结构化数据，避免两套 SQL 漂移）----

type feedingRow struct {
	T     time.Time
	Type  string
	Dur   int
	Amt   int
	Side  string
	Brand string
	Note  string
}

type diaperRow struct {
	T    time.Time
	Type string
	Note string
}

// spanRow 区间型记录（睡眠/户外）。Ongoing=true 时 End 为零值，「进行中」视为延伸到 now。
type spanRow struct {
	Start   time.Time
	End     time.Time
	Ongoing bool
	Note    string
}

type tempRow struct {
	T    time.Time
	Val  float64
	Loc  string
	Note string
}

type suppRow struct {
	T    time.Time
	Name string
	Val  float64
	Unit string
	Note string
}

// growthRow 成长记录。T 为纯日历日（YYYY-MM-DD），不做时区换算。
type growthRow struct {
	T        time.Time
	H, W, Hd float64
	Note     string
}

// reportData 一次导出的全部数据：宝宝档案 + 7 类记录 + 窗口。
type reportData struct {
	BabyID    int64
	BabyName  string
	Gender    string // male / female / 空（保密）
	BirthDate time.Time
	Owner     string // 导出人用户名
	TzOffset  int
	Days      int // 0 = 全部
	StartStr  string
	EndStr    string
	Generated time.Time // 用户时区

	Feeding    []feedingRow
	Diaper     []diaperRow
	Sleep      []spanRow
	Temp       []tempRow
	Outdoor    []spanRow
	Supplement []suppRow
	Growth     []growthRow
}

// queryReportData 按窗口（Days=0 为全部）取一份宝宝的全部记录。
// 睡眠/户外用「区间重叠」而非 started_at >= start：昨天 20:00 开始、
// 今天 06:00 结束的记录其今日部分落在窗口内；进行中视为延伸到无穷远。
func queryReportData(babyID, userID int64, tzOffset, days int) *reportData {
	d := &reportData{BabyID: babyID, TzOffset: tzOffset, Days: days}

	var birthDate string
	database.DB.QueryRow("SELECT name, gender, COALESCE(birth_date, '') FROM babies WHERE id = ?", babyID).Scan(&d.BabyName, &d.Gender, &birthDate)
	d.BirthDate = birthTimeFromDB(birthDate)
	database.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&d.Owner)
	loc := time.FixedZone("user", tzOffset*60)
	d.Generated = time.Now().In(loc)

	occurredFilter, occurredArgs := "", []interface{}{babyID}
	startedFilter, startedArgs := "", []interface{}{babyID}
	if days > 0 && days <= 365 {
		start, end := windowRangeUTC(tzOffset, days)
		d.StartStr, d.EndStr = start, end
		occurredFilter = " AND occurred_at >= ?"
		occurredArgs = append(occurredArgs, start)
		startedFilter = spanOverlapFilter()
		startedArgs = append(startedArgs, end, start)
	}

	// 喂奶
	if rs, err := database.DB.Query(
		"SELECT type, duration_minutes, amount_ml, side, brand, note, occurred_at FROM feeding_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var typ, side, brand, note, occ string
			var dur, amt int
			rs.Scan(&typ, &dur, &amt, &side, &brand, &note, &occ)
			d.Feeding = append(d.Feeding, feedingRow{parseTime(occ), typ, dur, amt, side, brand, note})
		}
		rs.Close()
	}

	// 尿布
	if rs, err := database.DB.Query(
		"SELECT type, note, occurred_at FROM diaper_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var typ, note, occ string
			rs.Scan(&typ, &note, &occ)
			d.Diaper = append(d.Diaper, diaperRow{parseTime(occ), typ, note})
		}
		rs.Close()
	}

	// 睡眠
	if rs, err := database.DB.Query(
		"SELECT started_at, COALESCE(ended_at, ''), note FROM sleep_records WHERE baby_id = ?"+startedFilter+" ORDER BY started_at DESC",
		startedArgs...,
	); err == nil {
		for rs.Next() {
			var st, en, note string
			rs.Scan(&st, &en, &note)
			r := spanRow{Start: parseTime(st), Note: note}
			if en != "" {
				r.End = parseTime(en)
			} else {
				r.Ongoing = true
			}
			d.Sleep = append(d.Sleep, r)
		}
		rs.Close()
	}

	// 体温
	if rs, err := database.DB.Query(
		"SELECT temperature, location, note, occurred_at FROM temperature_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var temp float64
			var loc, note, occ string
			rs.Scan(&temp, &loc, &note, &occ)
			d.Temp = append(d.Temp, tempRow{parseTime(occ), temp, loc, note})
		}
		rs.Close()
	}

	// 户外
	if rs, err := database.DB.Query(
		"SELECT started_at, COALESCE(ended_at, ''), note FROM outdoor_records WHERE baby_id = ?"+startedFilter+" ORDER BY started_at DESC",
		startedArgs...,
	); err == nil {
		for rs.Next() {
			var st, en, note string
			rs.Scan(&st, &en, &note)
			r := spanRow{Start: parseTime(st), Note: note}
			if en != "" {
				r.End = parseTime(en)
			} else {
				r.Ongoing = true
			}
			d.Outdoor = append(d.Outdoor, r)
		}
		rs.Close()
	}

	// 补剂
	if rs, err := database.DB.Query(
		"SELECT name, dosage_value, dosage_unit, note, occurred_at FROM supplement_records WHERE baby_id = ?"+occurredFilter+" ORDER BY occurred_at DESC",
		occurredArgs...,
	); err == nil {
		for rs.Next() {
			var name, unit, note, occ string
			var val float64
			rs.Scan(&name, &val, &unit, &note, &occ)
			d.Supplement = append(d.Supplement, suppRow{parseTime(occ), name, val, unit, note})
		}
		rs.Close()
	}

	// 成长（不受时间窗口限制，与 CSV 口径一致）
	if rs, err := database.DB.Query(
		"SELECT measured_at, height_cm, weight_kg, head_cm, note FROM growth_records WHERE baby_id = ? ORDER BY measured_at DESC",
		babyID,
	); err == nil {
		for rs.Next() {
			var occ, note string
			var h, w, hd float64
			rs.Scan(&occ, &h, &w, &hd, &note)
			d.Growth = append(d.Growth, growthRow{measuredTimeFromDB(occ), h, w, hd, note})
		}
		rs.Close()
	}

	return d
}