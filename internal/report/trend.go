package report

import (
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/util"
)

// TrendItem 趋势序列单点。
type TrendItem struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TrendSeries 趋势序列（按点位）。
type TrendSeries struct {
	PointCode string       `json:"point_code"`
	ParamType string       `json:"param_type"`
	Unit      string       `json:"unit"`
	Items     []TrendItem  `json:"items"`
}

// BuildTrendSeries 构造趋势序列。
func BuildTrendSeries(pointCode, paramType string, readings []model.Reading) TrendSeries {
	items := make([]TrendItem, 0, len(readings))
	for _, r := range readings {
		items = append(items, TrendItem{Timestamp: r.MeasuredAt, Value: r.Value})
	}
	return TrendSeries{
		PointCode: pointCode,
		ParamType: paramType,
		Unit:      model.ParamUnits[paramType],
		Items:     items,
	}
}

// Downsample 降采样：超过 max 点则均匀抽稀。
func Downsample(series TrendSeries, max int) TrendSeries {
	if max <= 0 || len(series.Items) <= max {
		return series
	}
	step := float64(len(series.Items)) / float64(max)
	out := make([]TrendItem, 0, max)
	for i := 0; i < max; i++ {
		idx := int(float64(i) * step)
		if idx >= len(series.Items) {
			idx = len(series.Items) - 1
		}
		out = append(out, series.Items[idx])
	}
	series.Items = out
	return series
}

// Bucket 时间桶。
type Bucket struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Avg   float64   `json:"avg"`
	Max   float64   `json:"max"`
	Count int       `json:"count"`
}

// Bucketize 按时长分桶聚合。
func Bucketize(readings []model.Reading, bucketDur time.Duration) []Bucket {
	if len(readings) == 0 {
		return nil
	}
	start := readings[0].MeasuredAt
	var buckets []Bucket
	var cur *Bucket
	for _, r := range readings {
		idx := int(r.MeasuredAt.Sub(start) / bucketDur)
		for len(buckets) <= idx {
			bStart := start.Add(time.Duration(len(buckets)) * bucketDur)
			buckets = append(buckets, Bucket{
				Start: bStart,
				End:   bStart.Add(bucketDur),
			})
		}
		cur = &buckets[idx]
		cur.Count++
		cur.Avg += r.Value
		if r.Value > cur.Max {
			cur.Max = r.Value
		}
	}
	for i := range buckets {
		if buckets[i].Count > 0 {
			buckets[i].Avg = util.Round(buckets[i].Avg/float64(buckets[i].Count), 2)
		}
	}
	return buckets
}