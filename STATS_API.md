# Stats API Implementation

## Endpoint
`GET /api/v1/stats`

## Authentication
Requires either:
- Bearer Token (JWT)
- X-API-Key header

## Response Format

```json
{
  "success": true,
  "message": "stats fetched successfully",
  "data": {
    "summary": {
      "totalExecutions": 12847,      // Total jobs in last 30 days
      "successRate": 94.2,            // Percentage (0-100) with 1 decimal
      "avgResponseTime": 342,         // Milliseconds (integer)
      "creditsRemaining": 8450,       // From User.credits
      "creditsChange": -150           // Negative value = consumed
    },
    "languageUsage": [
      {
        "language": "python",
        "count": 5234,
        "percentage": 41
      }
      // Top 4 languages, sorted by count descending
    ],
    "weeklyTrend": [
      {
        "day": "Mon",
        "date": "2025-12-15",
        "executions": 1834
      }
      // Last 7 days (today and previous 6)
    ],
    "insights": {
      "topLanguage": "python",
      "topLanguagePercentage": 41,
      "peakUsageDay": "Thursday",
      "peakUsageTime": "afternoons"
    }
  }
}
```

## Implementation Details

### Repository Functions (`internal/repository/stats.go`)

1. **GetTotalExecutions**: Counts all jobs for user in last 30 days
2. **GetSuccessRate**: Calculates percentage of successful executions (excludes SandboxError/InternalError)
3. **GetAvgResponseTime**: Average time between startedAt and finishedAt for successful jobs
4. **GetLanguageUsage**: Top 4 languages with counts and percentages
5. **GetWeeklyTrend**: Daily execution counts for last 7 days (fills in missing days with 0)
6. **GetInsights**: Derives insights from language usage and weekly trend data

### Handler (`internal/handler/stats/handler.go`)

- Fetches all stats data
- Formats response according to spec
- Returns 200 OK on success

### Routes (`internal/routes/stats.go`)

- Registered at `/api/v1/stats`
- Uses HybridAuthMiddleware for authentication

## Notes

- All date ranges use last 30 days from current time
- Success rate excludes jobs with SandboxError or InternalError
- Weekly trend always returns 7 days (fills missing days with 0 executions)
- Peak usage time is currently static ("afternoons") - can be enhanced with hourly aggregation
- Credits consumed calculation assumes 1 credit per execution (can be adjusted based on actual credit deduction logic)
