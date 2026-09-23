// Package data 提供《7岁以下儿童生长标准》(WS/T 423-2022) 的百分位数值表。
// 数据内嵌自 wst423_2022.json（转录自卫健委官方 PDF 表 A.1/A.3(A.4)/A.11/A.12，
// 已与官方数值多处交叉核对：男/女体重、女童身高 0 月，女童身高 5岁6月/6岁 中位数均一致）。
//
// 表结构说明：0-11 月逐月，1 岁起每 3 个月一行（源文档发布即为此密度），
// 头围仅 0-3 岁。月龄为整月或整岁；2 岁前为身长、2 岁后为身高。
// 与旧 WHO LMS 表不同，本表直接使用官方印制的 P3/P10/P25/P50/P75/P90/P97 七点，
// 百分位由相邻点分段线性插值得到。
package data

import (
	_ "embed"
	"encoding/json"
	"sort"
	"sync"
)

//go:embed wst423_2022.json
var rawJSON []byte

// PctRow 一行百分位数据（月龄 + 七个百分位点）
type PctRow struct {
	M   float64 `json:"m"`
	P3  float64 `json:"p3"`
	P10 float64 `json:"p10"`
	P25 float64 `json:"p25"`
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P90 float64 `json:"p90"`
	P97 float64 `json:"p97"`
}

var (
	once    sync.Once
	tables  map[string][]PctRow
	loadErr error
)

func load() {
	once.Do(func() {
		tables = map[string][]PctRow{}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rawJSON, &raw); err != nil {
			loadErr = err
			return
		}
		for k, v := range raw {
			if k == "meta" {
				continue
			}
			var rows []PctRow
			if err := json.Unmarshal(v, &rows); err != nil {
				loadErr = err
				return
			}
			tables[k] = rows
		}
	})
}

// Err 返回数据加载错误（nil 表示正常）
func Err() error {
	load()
	return loadErr
}

// Table 返回某性别(sex: male/female)某指标(metric: weight/height/head)的百分位行
func Table(metric, sex string) []PctRow {
	load()
	return tables[metric+"-"+sex]
}

// Percentiles 返回可用的百分位列名（与官方表一致）
func Percentiles() []string {
	return []string{"p3", "p10", "p25", "p50", "p75", "p90", "p97"}
}

func colAt(r *PctRow, col string) float64 {
	switch col {
	case "p3":
		return r.P3
	case "p10":
		return r.P10
	case "p25":
		return r.P25
	case "p50":
		return r.P50
	case "p75":
		return r.P75
	case "p90":
		return r.P90
	default:
		return r.P97
	}
}

// Value 返回某百分位列在月龄 month（可含小数）处的线性插值。
// 超出表格范围（0 或 最后一行之后）取端点值。
func Value(metric, sex, col string, month float64) float64 {
	rows := Table(metric, sex)
	if len(rows) == 0 || loadErr != nil {
		return 0
	}
	if month <= rows[0].M {
		return colAt(&rows[0], col)
	}
	last := rows[len(rows)-1]
	if month >= last.M {
		return colAt(&last, col)
	}
	hi := sort.Search(len(rows), func(i int) bool { return rows[i].M >= month })
	if hi >= len(rows) {
		return colAt(&last, col)
	}
	lo := hi - 1
	if lo < 0 {
		return colAt(&rows[0], col)
	}
	a, b := rows[lo], rows[hi]
	if b.M == a.M {
		return colAt(&a, col)
	}
	t := (month - a.M) / (b.M - a.M)
	return colAt(&a, col) + (colAt(&b, col)-colAt(&a, col))*t
}

var psum = []float64{3, 10, 25, 50, 75, 90, 97}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Percentile 返回月龄 month 处数值 value 对应的百分位（0-100）。
// 依据官方印制的 P3/P10/P25/P50/P75/P90/P97 七点分段线性插值；
// 超出两端时沿最近两点斜率外推并收敛到 [0.5, 99.5]。
func Percentile(metric, sex string, month, value float64) float64 {
	if value <= 0 {
		return 0
	}
	cols := Percentiles()
	vs := make([]float64, len(cols))
	for i, c := range cols {
		vs[i] = Value(metric, sex, c, month)
	}
	// 低于 P3：沿 P3→P10 斜率外推
	if value <= vs[0] {
		if vs[1] <= vs[0] {
			return 1
		}
		slope := (vs[1] - vs[0]) / (psum[1] - psum[0])
		return clampFloat(psum[0]-(vs[0]-value)/slope, 0.5, psum[0])
	}
	// 高于 P97
	if value >= vs[len(vs)-1] {
		if vs[len(vs)-1] <= vs[len(vs)-2] {
			return 99
		}
		slope := (vs[len(vs)-1] - vs[len(vs)-2]) / (psum[len(psum)-1] - psum[len(psum)-2])
		return clampFloat(psum[len(psum)-1]+(value-vs[len(vs)-1])/slope, psum[len(psum)-1], 99.5)
	}
	for i := 0; i < len(vs)-1; i++ {
		if value >= vs[i] && value <= vs[i+1] {
			if vs[i+1] == vs[i] {
				return psum[i]
			}
			frac := (value - vs[i]) / (vs[i+1] - vs[i])
			return psum[i] + (psum[i+1]-psum[i])*frac
		}
	}
	return 50
}