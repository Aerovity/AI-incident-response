package analytics

import (
	"fmt"
	"incident-ai/models"
	"sort"
	"time"
)

// TrendPoint represents a data point in a trend
type TrendPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Count     int                    `json:"count"`
	AvgTime   float64                `json:"avg_resolution_time_seconds,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Trend represents a time-series trend
type Trend struct {
	Label      string       `json:"label"`
	Points     []TrendPoint `json:"points"`
	Direction  string       `json:"direction"` // "increasing", "decreasing", "stable"
	ChangeRate float64      `json:"change_rate_percent"`
}

// AnalyticsReport contains comprehensive analytics data
type AnalyticsReport struct {
	// Summary Stats
	TotalIncidents    int     `json:"total_incidents"`
	ResolvedIncidents int     `json:"resolved_incidents"`
	FailedIncidents   int     `json:"failed_incidents"`
	SuccessRate       float64 `json:"success_rate_percent"`

	// Time-based Analytics
	AvgResolutionTime    float64            `json:"avg_resolution_time_seconds"`
	MedianResolutionTime float64            `json:"median_resolution_time_seconds"`
	FastestResolution    float64            `json:"fastest_resolution_seconds"`
	SlowestResolution    float64            `json:"slowest_resolution_seconds"`

	// Incident Breakdown
	IncidentsByType     map[string]int     `json:"incidents_by_type"`
	IncidentsByPriority map[string]int     `json:"incidents_by_priority"`
	IncidentsByStatus   map[string]int     `json:"incidents_by_status"`

	// Trends
	HourlyTrend         Trend              `json:"hourly_trend"`
	DailyTrend          Trend              `json:"daily_trend"`
	TypeTrends          map[string]Trend   `json:"type_trends"`

	// Performance Metrics
	CachedFixUsage      int                `json:"cached_fix_usage_count"`
	CachedFixRate       float64            `json:"cached_fix_rate_percent"`
	AICallsMade         int                `json:"ai_calls_made"`

	// Hot Spots
	MostFrequentType    string             `json:"most_frequent_incident_type"`
	MostProblematicType string             `json:"most_problematic_type"` // Highest failure rate

	// Time Distribution
	TimeDistribution    map[string]int     `json:"time_distribution"` // Hour of day -> count

	// Recent Activity
	Last24Hours         int                `json:"incidents_last_24h"`
	Last7Days           int                `json:"incidents_last_7d"`
	Last30Days          int                `json:"incidents_last_30d"`

	// Predictions (Simple)
	PredictedNextHour   float64            `json:"predicted_incidents_next_hour"`
	TrendStatus         string             `json:"trend_status"` // "improving", "degrading", "stable"
}

// Engine performs analytics calculations
type Engine struct{}

// NewEngine creates a new analytics engine
func NewEngine() *Engine {
	return &Engine{}
}

// GenerateReport creates a comprehensive analytics report
func (e *Engine) GenerateReport(incidents []*models.Incident) *AnalyticsReport {
	report := &AnalyticsReport{
		IncidentsByType:     make(map[string]int),
		IncidentsByPriority: make(map[string]int),
		IncidentsByStatus:   make(map[string]int),
		TimeDistribution:    make(map[string]int),
		TypeTrends:          make(map[string]Trend),
	}

	if len(incidents) == 0 {
		return report
	}

	// Basic counts
	report.TotalIncidents = len(incidents)

	var resolutionTimes []float64
	var totalResolutionTime float64

	for _, inc := range incidents {
		// Count by status
		report.IncidentsByStatus[string(inc.Status)]++

		if inc.Status == models.StatusResolved {
			report.ResolvedIncidents++
			if inc.ResolvedAt != nil {
				resTime := inc.ResolvedAt.Sub(inc.DetectedAt).Seconds()
				resolutionTimes = append(resolutionTimes, resTime)
				totalResolutionTime += resTime
			}
		} else if inc.Status == models.StatusFailed {
			report.FailedIncidents++
		}

		// Count by type
		report.IncidentsByType[string(inc.Type)]++

		// Count by priority
		report.IncidentsByPriority[string(inc.Priority)]++

		// Time distribution (hour of day)
		hour := inc.DetectedAt.Hour()
		report.TimeDistribution[formatHour(hour)]++

		// Cached fix usage
		if inc.UsedCachedFix {
			report.CachedFixUsage++
		} else if inc.Status == models.StatusResolved {
			report.AICallsMade++
		}
	}

	// Calculate success rate
	if report.TotalIncidents > 0 {
		report.SuccessRate = (float64(report.ResolvedIncidents) / float64(report.TotalIncidents)) * 100
	}

	// Calculate cached fix rate
	if report.ResolvedIncidents > 0 {
		report.CachedFixRate = (float64(report.CachedFixUsage) / float64(report.ResolvedIncidents)) * 100
	}

	// Resolution time statistics
	if len(resolutionTimes) > 0 {
		report.AvgResolutionTime = totalResolutionTime / float64(len(resolutionTimes))
		report.MedianResolutionTime = calculateMedian(resolutionTimes)
		report.FastestResolution = calculateMin(resolutionTimes)
		report.SlowestResolution = calculateMax(resolutionTimes)
	}

	// Find hot spots
	report.MostFrequentType = findMaxKey(report.IncidentsByType)
	report.MostProblematicType = e.findMostProblematic(incidents)

	// Time-based counts
	now := time.Now()
	report.Last24Hours = e.countSince(incidents, now.Add(-24*time.Hour))
	report.Last7Days = e.countSince(incidents, now.Add(-7*24*time.Hour))
	report.Last30Days = e.countSince(incidents, now.Add(-30*24*time.Hour))

	// Generate trends
	report.HourlyTrend = e.generateHourlyTrend(incidents)
	report.DailyTrend = e.generateDailyTrend(incidents)
	report.TypeTrends = e.generateTypeTrends(incidents)

	// Predictions
	report.PredictedNextHour = e.predictNextHour(incidents)
	report.TrendStatus = e.determineTrendStatus(report.HourlyTrend)

	return report
}

// generateHourlyTrend creates an hourly incident trend
func (e *Engine) generateHourlyTrend(incidents []*models.Incident) Trend {
	points := make(map[time.Time]*TrendPoint)

	for _, inc := range incidents {
		// Round to hour
		hourTime := inc.DetectedAt.Truncate(time.Hour)

		if points[hourTime] == nil {
			points[hourTime] = &TrendPoint{
				Timestamp: hourTime,
				Count:     0,
			}
		}
		points[hourTime].Count++
	}

	// Convert to slice and sort
	var pointSlice []TrendPoint
	for _, p := range points {
		pointSlice = append(pointSlice, *p)
	}
	sort.Slice(pointSlice, func(i, j int) bool {
		return pointSlice[i].Timestamp.Before(pointSlice[j].Timestamp)
	})

	trend := Trend{
		Label:  "Incidents per Hour",
		Points: pointSlice,
	}

	// Calculate direction and rate
	if len(pointSlice) >= 2 {
		firstHalf := pointSlice[:len(pointSlice)/2]
		secondHalf := pointSlice[len(pointSlice)/2:]

		avgFirst := calculateAvgCount(firstHalf)
		avgSecond := calculateAvgCount(secondHalf)

		if avgSecond > avgFirst*1.1 {
			trend.Direction = "increasing"
			trend.ChangeRate = ((avgSecond - avgFirst) / avgFirst) * 100
		} else if avgSecond < avgFirst*0.9 {
			trend.Direction = "decreasing"
			trend.ChangeRate = ((avgFirst - avgSecond) / avgFirst) * 100
		} else {
			trend.Direction = "stable"
			trend.ChangeRate = 0
		}
	} else {
		trend.Direction = "stable"
	}

	return trend
}

// generateDailyTrend creates a daily incident trend
func (e *Engine) generateDailyTrend(incidents []*models.Incident) Trend {
	points := make(map[time.Time]*TrendPoint)

	for _, inc := range incidents {
		// Round to day
		dayTime := time.Date(inc.DetectedAt.Year(), inc.DetectedAt.Month(),
			inc.DetectedAt.Day(), 0, 0, 0, 0, inc.DetectedAt.Location())

		if points[dayTime] == nil {
			points[dayTime] = &TrendPoint{
				Timestamp: dayTime,
				Count:     0,
			}
		}
		points[dayTime].Count++
	}

	// Convert to slice and sort
	var pointSlice []TrendPoint
	for _, p := range points {
		pointSlice = append(pointSlice, *p)
	}
	sort.Slice(pointSlice, func(i, j int) bool {
		return pointSlice[i].Timestamp.Before(pointSlice[j].Timestamp)
	})

	trend := Trend{
		Label:  "Incidents per Day",
		Points: pointSlice,
	}

	// Calculate direction
	if len(pointSlice) >= 2 {
		recent := pointSlice[len(pointSlice)-1].Count
		previous := pointSlice[len(pointSlice)-2].Count

		if recent > previous {
			trend.Direction = "increasing"
			trend.ChangeRate = float64(recent-previous) / float64(previous) * 100
		} else if recent < previous {
			trend.Direction = "decreasing"
			trend.ChangeRate = float64(previous-recent) / float64(previous) * 100
		} else {
			trend.Direction = "stable"
		}
	}

	return trend
}

// generateTypeTrends creates trends for each incident type
func (e *Engine) generateTypeTrends(incidents []*models.Incident) map[string]Trend {
	typeData := make(map[string][]TrendPoint)

	for _, inc := range incidents {
		incType := string(inc.Type)
		hourTime := inc.DetectedAt.Truncate(time.Hour)

		// Find or create point for this hour
		var found bool
		for i := range typeData[incType] {
			if typeData[incType][i].Timestamp.Equal(hourTime) {
				typeData[incType][i].Count++
				found = true
				break
			}
		}
		if !found {
			typeData[incType] = append(typeData[incType], TrendPoint{
				Timestamp: hourTime,
				Count:     1,
			})
		}
	}

	trends := make(map[string]Trend)
	for incType, points := range typeData {
		sort.Slice(points, func(i, j int) bool {
			return points[i].Timestamp.Before(points[j].Timestamp)
		})

		trend := Trend{
			Label:     incType,
			Points:    points,
			Direction: "stable",
		}

		if len(points) >= 2 {
			first := points[0].Count
			last := points[len(points)-1].Count

			if last > first {
				trend.Direction = "increasing"
				trend.ChangeRate = float64(last-first) / float64(first) * 100
			} else if last < first {
				trend.Direction = "decreasing"
				trend.ChangeRate = float64(first-last) / float64(first) * 100
			}
		}

		trends[incType] = trend
	}

	return trends
}

// predictNextHour uses simple moving average to predict next hour
func (e *Engine) predictNextHour(incidents []*models.Incident) float64 {
	if len(incidents) == 0 {
		return 0
	}

	// Count incidents per hour for last 24 hours
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)

	hourlyCounts := make(map[int]int)
	for _, inc := range incidents {
		if inc.DetectedAt.After(cutoff) {
			hour := inc.DetectedAt.Hour()
			hourlyCounts[hour]++
		}
	}

	// Calculate average
	if len(hourlyCounts) == 0 {
		return 0
	}

	total := 0
	for _, count := range hourlyCounts {
		total += count
	}

	return float64(total) / float64(len(hourlyCounts))
}

// determineTrendStatus determines overall system health trend
func (e *Engine) determineTrendStatus(hourlyTrend Trend) string {
	if hourlyTrend.Direction == "increasing" && hourlyTrend.ChangeRate > 20 {
		return "degrading"
	} else if hourlyTrend.Direction == "decreasing" && hourlyTrend.ChangeRate > 20 {
		return "improving"
	}
	return "stable"
}

// Helper functions

func (e *Engine) countSince(incidents []*models.Incident, since time.Time) int {
	count := 0
	for _, inc := range incidents {
		if inc.DetectedAt.After(since) {
			count++
		}
	}
	return count
}

func (e *Engine) findMostProblematic(incidents []*models.Incident) string {
	typeFailures := make(map[string]int)
	typeTotals := make(map[string]int)

	for _, inc := range incidents {
		incType := string(inc.Type)
		typeTotals[incType]++
		if inc.Status == models.StatusFailed {
			typeFailures[incType]++
		}
	}

	maxRate := 0.0
	maxType := ""

	for incType, total := range typeTotals {
		if total > 0 {
			failureRate := float64(typeFailures[incType]) / float64(total)
			if failureRate > maxRate {
				maxRate = failureRate
				maxType = incType
			}
		}
	}

	return maxType
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func findMaxKey(m map[string]int) string {
	max := 0
	maxKey := ""
	for k, v := range m {
		if v > max {
			max = v
			maxKey = k
		}
	}
	return maxKey
}

func calculateAvgCount(points []TrendPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	total := 0
	for _, p := range points {
		total += p.Count
	}
	return float64(total) / float64(len(points))
}

func formatHour(hour int) string {
	if hour == 0 {
		return "12 AM"
	} else if hour < 12 {
		return fmt.Sprintf("%d AM", hour)
	} else if hour == 12 {
		return "12 PM"
	} else {
		return fmt.Sprintf("%d PM", hour-12)
	}
}
