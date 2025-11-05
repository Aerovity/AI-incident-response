# 📊 Advanced Analytics & Trending System

## Overview

The incident response system now includes a comprehensive analytics and trending engine that provides real-time insights into incident patterns, resolution performance, and predictive forecasting.

## Accessing Analytics

```bash
# Start the system
go run main.go

# In another terminal, get analytics
curl http://localhost:8080/analytics
```

## Features

### 1. **Summary Statistics**
- Total incidents handled (all time)
- Resolved vs failed incidents
- Success rate percentage
- Resolution time metrics:
  - Average resolution time
  - Median resolution time
  - Fastest resolution
  - Slowest resolution

### 2. **Performance Metrics**
- **Cached Fix Usage**: How many incidents were resolved using learned fixes (no AI call needed)
- **Cached Fix Rate**: Percentage of incidents resolved instantly via cache
- **AI Calls Made**: Number of times Cloudflare AI was invoked
- **Efficiency Tracking**: Monitor how the system learns and improves over time

### 3. **Incident Breakdown**
- **By Type**: Count of each incident type (SERVICE_DOWN, CONFIG_ERROR, etc.)
- **By Priority**: Distribution across Critical/High/Medium/Low
- **By Status**: Currently pending, analyzing, fixing, resolved, or failed

### 4. **Time-Based Trends**

#### Hourly Trend
- Incidents per hour over time
- Direction: increasing, decreasing, or stable
- Change rate percentage
- Data points with timestamps

#### Daily Trend
- Incidents per day
- Day-over-day comparison
- Long-term pattern identification

#### Type-Specific Trends
- Individual trends for each incident type
- Identify which types are becoming more/less frequent
- Spot emerging issues early

### 5. **Hot Spots & Problem Areas**
- **Most Frequent Type**: Which incident happens most often
- **Most Problematic Type**: Which type has the highest failure rate
- **Time Distribution**: When incidents occur (hour of day analysis)
  - Identify peak incident hours
  - Plan maintenance windows
  - Optimize staffing

### 6. **Recent Activity Windows**
- Last 24 hours: Short-term activity
- Last 7 days: Weekly trend
- Last 30 days: Monthly overview

### 7. **Predictions**
- **Next Hour Forecast**: Predicted number of incidents in the next hour
  - Uses simple moving average
  - Based on last 24 hours of data
- **Trend Status**: Overall system health direction
  - "improving": Incidents decreasing significantly (>20%)
  - "degrading": Incidents increasing significantly (>20%)
  - "stable": No major change

## Example Use Cases

### 1. Daily Standup Report
```bash
curl http://localhost:8080/analytics | jq '{
  last_24h: .incidents_last_24h,
  success_rate: .success_rate_percent,
  trend: .trend_status,
  most_common: .most_frequent_incident_type
}'
```

### 2. Performance Monitoring
```bash
curl http://localhost:8080/analytics | jq '{
  avg_resolution: .avg_resolution_time_seconds,
  cached_fix_rate: .cached_fix_rate_percent,
  ai_calls: .ai_calls_made
}'
```

### 3. Capacity Planning
```bash
curl http://localhost:8080/analytics | jq '{
  prediction: .predicted_incidents_next_hour,
  peak_hours: .time_distribution,
  trend: .hourly_trend.direction
}'
```

### 4. Problem Identification
```bash
curl http://localhost:8080/analytics | jq '{
  most_frequent: .most_frequent_incident_type,
  most_problematic: .most_problematic_type,
  failed: .failed_incidents
}'
```

## Analytics Architecture

```
┌─────────────┐
│   Memory    │
│    Store    │◄──── Stores all incidents with metadata
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Analytics  │
│   Engine    │◄──── Processes incident data
└──────┬──────┘
       │
       ├──► Trend Analysis (hourly/daily/by-type)
       ├──► Statistical Calculations
       ├──► Hot Spot Detection
       ├──► Prediction Generation
       │
       ▼
┌─────────────┐
│   HTTP      │
│  Endpoint   │◄──── Serves JSON reports
└─────────────┘
```

## Data Flow

1. **Incident Occurs** → Detected by monitor
2. **Incident Resolved** → Stored in memory with timestamps
3. **Analytics Request** → Engine pulls all incidents from store
4. **Processing** → Calculates trends, stats, predictions
5. **Response** → Returns comprehensive JSON report

## Key Algorithms

### Trend Direction Detection
- Compares first half vs second half of data points
- >10% increase = "increasing"
- >10% decrease = "decreasing"
- Within 10% = "stable"

### Median Resolution Time
- Sorts all resolution times
- Returns middle value (or average of two middle values)
- More robust than mean for outliers

### Prediction (Next Hour)
- Simple Moving Average
- Counts incidents per hour in last 24 hours
- Averages to predict next hour
- Future: Could implement exponential smoothing or ML models

### Most Problematic Type
- Calculates failure rate per type (failures / total)
- Returns type with highest failure rate
- Helps prioritize fix improvements

## Interpreting Results

### Success Rate
- **>95%**: Excellent - system is very reliable
- **85-95%**: Good - minor improvements needed
- **70-85%**: Fair - investigate failure patterns
- **<70%**: Poor - serious issues need attention

### Cached Fix Rate
- **>50%**: System is learning well
- **25-50%**: Learning is occurring but slower
- **<25%**: Many new/unique incidents

### Trend Status
- **"improving"**: Incidents decreasing - system becoming more stable
- **"stable"**: Consistent incident rate - monitor closely
- **"degrading"**: Incidents increasing - investigate root causes

### Time Distribution
- High counts at specific hours → scheduled jobs causing issues
- Evenly distributed → random failures
- Clustering → external dependency patterns

## Tips for Analysis

1. **Run analytics periodically** to track changes over time
2. **Compare cached fix rate** over days/weeks to measure learning
3. **Watch for trend changes** that indicate new problems
4. **Correlate time distribution** with deployment schedules
5. **Use predictions** for capacity planning and staffing

## Future Enhancements

- [ ] Export analytics to time-series database (Prometheus, InfluxDB)
- [ ] Anomaly detection using statistical methods
- [ ] Machine learning-based predictions
- [ ] Correlation analysis between incident types
- [ ] SLA tracking and alerting
- [ ] Dashboard web UI with charts
- [ ] Historical comparison (week-over-week, month-over-month)
- [ ] Alert fatigue analysis
- [ ] MTTR (Mean Time To Resolution) tracking by type

## API Reference

### GET /analytics

**Response Format:**
```json
{
  "total_incidents": int,
  "resolved_incidents": int,
  "failed_incidents": int,
  "success_rate_percent": float64,
  "avg_resolution_time_seconds": float64,
  "median_resolution_time_seconds": float64,
  "fastest_resolution_seconds": float64,
  "slowest_resolution_seconds": float64,
  "incidents_by_type": {"type": count},
  "incidents_by_priority": {"priority": count},
  "incidents_by_status": {"status": count},
  "hourly_trend": {
    "label": string,
    "points": [{"timestamp": time, "count": int}],
    "direction": string,
    "change_rate_percent": float64
  },
  "daily_trend": {...},
  "type_trends": {"type": {...}},
  "cached_fix_usage_count": int,
  "cached_fix_rate_percent": float64,
  "ai_calls_made": int,
  "most_frequent_incident_type": string,
  "most_problematic_type": string,
  "time_distribution": {"hour": count},
  "incidents_last_24h": int,
  "incidents_last_7d": int,
  "incidents_last_30d": int,
  "predicted_incidents_next_hour": float64,
  "trend_status": string
}
```

**HTTP Status Codes:**
- `200 OK`: Analytics generated successfully
- `500 Internal Server Error`: Error generating analytics

## Integration Examples

### Shell Script for Daily Reports
```bash
#!/bin/bash
# daily-report.sh

echo "=== Incident Response Daily Report ==="
echo ""

ANALYTICS=$(curl -s http://localhost:8080/analytics)

echo "Total Incidents (24h): $(echo $ANALYTICS | jq -r '.incidents_last_24h')"
echo "Success Rate: $(echo $ANALYTICS | jq -r '.success_rate_percent')%"
echo "Avg Resolution: $(echo $ANALYTICS | jq -r '.avg_resolution_time_seconds')s"
echo "Trend: $(echo $ANALYTICS | jq -r '.trend_status')"
echo ""
echo "Most Common: $(echo $ANALYTICS | jq -r '.most_frequent_incident_type')"
echo "Most Problems: $(echo $ANALYTICS | jq -r '.most_problematic_type')"
```

### Python Script for Monitoring
```python
import requests
import json

def get_analytics():
    response = requests.get('http://localhost:8080/analytics')
    return response.json()

def check_health():
    analytics = get_analytics()

    # Alert if success rate drops below 80%
    if analytics['success_rate_percent'] < 80:
        print(f"⚠️ WARNING: Success rate is {analytics['success_rate_percent']}%")

    # Alert if degrading trend
    if analytics['trend_status'] == 'degrading':
        print("⚠️ WARNING: Incident trend is degrading")

    # Alert if many recent incidents
    if analytics['incidents_last_24h'] > 50:
        print(f"⚠️ WARNING: High incident volume: {analytics['incidents_last_24h']} in 24h")

if __name__ == '__main__':
    check_health()
```

## Troubleshooting

### No data in analytics
- Trigger some incidents first: `curl "http://localhost:8080/trigger-incident?type=crash"`
- Wait for them to be resolved
- Then check analytics

### Predictions show 0
- Need at least a few incidents in last 24 hours
- Prediction accuracy improves with more data

### Trends show "stable" always
- Need multiple data points over time
- Run system longer or trigger more varied incidents

---

**Built with Go + Cloudflare AI**
Part of the AI-Powered Incident Response System
