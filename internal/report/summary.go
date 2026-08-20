package report

import (
	"time"

	"cleanroom-monitor/internal/util"
)

// DailySummary 单日汇总。
type DailySummary struct {
	Date          string  `json:"date"`
	ReadingCount  int     `json:"reading_count"`
	OutOfRange    int     `json:"out_of_range"`
	CompliancePct float64 `json:"compliance_pct"`
	OpenAlerts    int     `json:"open_alerts"`
}

// BuildDailySummary 由统计值构造单日汇总。
func BuildDailySummary(day time.Time, readingCount, outOfRange, openAlerts int) DailySummary {
	pct := 0.0
	if readingCount > 0 {
		pct = float64(readingCount-outOfRange) / float64(readingCount) * 100
	}
	return DailySummary{
		Date:          util.FormatDate(day),
		ReadingCount:  readingCount,
		OutOfRange:    outOfRange,
		CompliancePct: util.Round(pct, 1),
		OpenAlerts:    openAlerts,
	}
}

// TrendSummary 阶段趋势汇总。
type TrendSummary struct {
	From          string  `json:"from"`
	To            string  `json:"to"`
	Days          int     `json:"days"`
	AvgCompliance float64 `json:"avg_compliance"`
}

// BuildTrendSummary 汇总多日达标率。
func BuildTrendSummary(from, to time.Time, daily []DailySummary) TrendSummary {
	if len(daily) == 0 {
		return TrendSummary{
			From: util.FormatDate(from),
			To:   util.FormatDate(to),
			Days: 0,
		}
	}
	sum := 0.0
	for _, d := range daily {
		sum += d.CompliancePct
	}
	return TrendSummary{
		From:          util.FormatDate(from),
		To:            util.FormatDate(to),
		Days:          len(daily),
		AvgCompliance: util.Round(sum/float64(len(daily)), 1),
	}
}