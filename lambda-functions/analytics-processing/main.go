package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/aggregator"
	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/converter"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ddbattr "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	dynamoClient   *dynamodb.Client
	s3Client       *s3.Client
	tableReadings  string
	tableAnalytics string
	s3Bucket       string
	defaultCtx     = context.Background()
)

type Reading struct {
	FacilityID string  `dynamodbav:"facilityId"`
	MeterID    string  `dynamodbav:"meterId"`
	Timestamp  int64   `dynamodbav:"timestamp"`
	Voltage    float64 `dynamodbav:"voltage"`
	Current    float64 `dynamodbav:"current"`
	PowerKW    float64 `dynamodbav:"powerKw"`
}

type HourlyData struct {
	Count      int     `json:"count"`
	TotalPower float64 `json:"total_power"`
	AvgPower   float64 `json:"avg_power"`
	MaxPower   float64 `json:"max_power"`
}

type DailyAnalytics struct {
	Date                string                `json:"date"`
	ReadingCount        int                   `json:"reading_count"`
	TotalConsumption    float64               `json:"total_consumption"`
	TotalConsumptionMWh float64               `json:"total_consumption_mwh"`
	AveragePower        float64               `json:"average_power"`
	PeakPower           float64               `json:"peak_power"`
	MinPower            float64               `json:"min_power"`
	MovingAverage       []float64             `json:"moving_average"`
	EstimatedCost       float64               `json:"estimated_cost"`
	CostBreakdown       map[string]float64    `json:"cost_breakdown"`
	AvgVoltage          float64               `json:"avg_voltage"`
	VoltageStdDev       float64               `json:"voltage_stddev"`
	AvgCurrent          float64               `json:"avg_current"`
	PowerFactor         float64               `json:"power_factor"`
	PeakHour            string                `json:"peak_hour"`
	HourlyData          map[string]HourlyData `json:"hourly_data"`
	CreatedAt           int64                 `dynamodbav:"createdAt" json:"created_at"`
}

type LambdaEvent struct {
	Date       string `json:"date"`        // YYYY-MM-DD (optional; defaults to yesterday)
	FacilityID string `json:"facility_id"` // optional; defaults to facility-001
}

type LambdaResponse struct {
	StatusCode int                    `json:"statusCode"`
	Body       map[string]interface{} `json:"body"`
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func init() {
	// Let the SDK discover region from environment/role; no need to force-set AWS_REGION.
	cfg, err := config.LoadDefaultConfig(defaultCtx)
	if err != nil {
		panic(fmt.Sprintf("unable to load AWS SDK config: %v", err))
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)

	// Env-driven names with safe defaults
	tableReadings = getenv("DDB_TABLE_READINGS", "EnergyReadings")
	tableAnalytics = getenv("DDB_TABLE_ANALYTICS", "AnalyticsSummaries")
	s3Bucket = getenv("S3_BUCKET", "energy-grid-reports")

	fmt.Printf("Cold start: ReadingsTable=%s AnalyticsTable=%s S3Bucket=%s\n",
		tableReadings, tableAnalytics, s3Bucket)
}

func Handler(ctx context.Context, event LambdaEvent) (LambdaResponse, error) {
	date := event.Date
	if date == "" {
		date = time.Now().AddDate(0, 0, -1).Format("2006-01-02") // default: yesterday
	}
	facilityID := event.FacilityID
	if facilityID == "" {
		facilityID = "facility-001"
	}

	fmt.Printf("Start daily aggregation: facility=%s date=%s\n", facilityID, date)

	readings, err := getReadingsForDate(ctx, facilityID, date, 2000) // sensible cap; paginate if needed
	if err != nil {
		return fail(500, err)
	}
	if len(readings) == 0 {
		return ok(map[string]interface{}{
			"message": "No data to process",
			"date":    date,
		})
	}

	analytics := calculateDailyAnalytics(readings, date)

	if err := storeAnalyticsSummary(ctx, facilityID, analytics); err != nil {
		// Non-fatal: continue to S3 report so the day isn’t lost
		fmt.Printf("WARN storeAnalyticsSummary: %v\n", err)
	}

	reportURL, err := generateReport(ctx, facilityID, date, analytics)
	if err != nil {
		fmt.Printf("WARN generateReport: %v\n", err)
	}

	return ok(map[string]interface{}{
		"message":    "Analytics processed successfully",
		"date":       date,
		"analytics":  analytics,
		"report_url": reportURL,
	})
}

func ok(body map[string]interface{}) (LambdaResponse, error) {
	return LambdaResponse{StatusCode: 200, Body: body}, nil
}

func fail(code int, err error) (LambdaResponse, error) {
	return LambdaResponse{
		StatusCode: code,
		Body:       map[string]interface{}{"error": err.Error()},
	}, err
}

// --- Data access ---

// getReadingsForDate queries all readings for the facility within the day, handling pagination.
func getReadingsForDate(ctx context.Context, facilityID, date string, pageLimit int32) ([]Reading, error) {
	startOfDay, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("bad date format %q: %w", date, err)
	}
	endOfDay := startOfDay.Add(24 * time.Hour)

	startTS := startOfDay.Unix()
	endTS := endOfDay.Unix()

	var (
		all       []Reading
		exclusive map[string]types.AttributeValue
		pageCount int
		maxPages  = 50 // guardrail
	)

	for {
		in := &dynamodb.QueryInput{
			TableName:              aws.String(tableReadings),
			KeyConditionExpression: aws.String("facilityId = :fid AND #ts BETWEEN :start AND :end"),
			ExpressionAttributeNames: map[string]string{
				"#ts": "timestamp",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":fid":   &types.AttributeValueMemberS{Value: facilityID},
				":start": &types.AttributeValueMemberN{Value: strconv.FormatInt(startTS, 10)},
				":end":   &types.AttributeValueMemberN{Value: strconv.FormatInt(endTS, 10)},
			},
			ScanIndexForward:  aws.Bool(true), // oldest -> newest
			Limit:             aws.Int32(pageLimit),
			ExclusiveStartKey: exclusive,
		}

		out, err := dynamoClient.Query(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("dynamodb query failed: %w", err)
		}

		var page []Reading
		if err := ddbattr.UnmarshalListOfMaps(out.Items, &page); err != nil {
			return nil, fmt.Errorf("unmarshal readings failed: %w", err)
		}
		all = append(all, page...)

		if len(out.LastEvaluatedKey) == 0 {
			break
		}
		exclusive = out.LastEvaluatedKey
		pageCount++
		if pageCount > maxPages {
			fmt.Printf("WARN: pagination stopped at %d pages (%d readings)\n", pageCount, len(all))
			break
		}
	}

	return all, nil
}

// --- Analytics ---

func calculateDailyAnalytics(readings []Reading, date string) DailyAnalytics {
	points := make([]aggregator.Point, len(readings))
	for i, r := range readings {
		points[i] = aggregator.Point{Value: r.PowerKW, Timestamp: time.Unix(r.Timestamp, 0)}
	}

	totalPower := aggregator.Sum(points)
	avgPower := safeAverage(points)
	movingAvg := aggregator.MovingAverage(points, 12) // configurable if needed

	conv := &converter.EnergyConverter{}
	totalConsumptionMWh := conv.KWhToMWh(totalPower)

	// Simple cost model—tune as needed
	peakCost := conv.CalculateCost(totalPower*0.4, 0.20, "peak")
	offPeakCost := conv.CalculateCost(totalPower*0.6, 0.20, "offpeak")
	totalCost := peakCost + offPeakCost

	peak, min := findMaxMin(points)
	hourly := calculateHourlyData(readings)
	peakHour := derivePeakHour(hourly)

	avgV := averageFloat(func(i int) float64 { return readings[i].Voltage }, len(readings))
	avgI := averageFloat(func(i int) float64 { return readings[i].Current }, len(readings))
	voltageStd := stddevFloat(func(i int) float64 { return readings[i].Voltage }, len(readings), avgV)

	// Apparent power ≈ V * I (average); avoid negative/NaN
	apparent := max0(avgV * avgI)
	powerFactor := 0.0
	if apparent > 0 {
		powerFactor = conv.CalculateEfficiency(apparent, avgPower)
	}

	return DailyAnalytics{
		Date:                date,
		ReadingCount:        len(readings),
		TotalConsumption:    round2(totalPower),
		TotalConsumptionMWh: round3(totalConsumptionMWh),
		AveragePower:        round2(avgPower),
		PeakPower:           round2(peak),
		MinPower:            round2(min),
		MovingAverage:       roundSlice(movingAvg, 2),
		EstimatedCost:       round2(totalCost),
		CostBreakdown: map[string]float64{
			"peak":    round2(peakCost),
			"offpeak": round2(offPeakCost),
		},
		AvgVoltage:    round2(avgV),
		VoltageStdDev: round3(voltageStd),
		AvgCurrent:    round2(avgI),
		PowerFactor:   round3(powerFactor),
		PeakHour:      peakHour,
		HourlyData:    hourly,
		CreatedAt:     time.Now().Unix(),
	}
}

func safeAverage(points []aggregator.Point) float64 {
	if len(points) == 0 {
		return 0
	}
	return aggregator.Average(points)
}

func findMaxMin(points []aggregator.Point) (maxVal, minVal float64) {
	if len(points) == 0 {
		return 0, 0
	}
	maxVal = points[0].Value
	minVal = points[0].Value
	for _, p := range points[1:] {
		if p.Value > maxVal {
			maxVal = p.Value
		}
		if p.Value < minVal {
			minVal = p.Value
		}
	}
	return
}

func calculateHourlyData(readings []Reading) map[string]HourlyData {
	hourly := make(map[string]HourlyData, 24)
	for _, r := range readings {
		h := time.Unix(r.Timestamp, 0).Format("15") // "00".."23"
		data := hourly[h]
		data.Count++
		data.TotalPower += r.PowerKW
		if r.PowerKW > data.MaxPower {
			data.MaxPower = r.PowerKW
		}
		hourly[h] = data
	}
	for h, d := range hourly {
		if d.Count > 0 {
			d.AvgPower = d.TotalPower / float64(d.Count)
			hourly[h] = d
		}
	}
	return hourly
}

func derivePeakHour(hourly map[string]HourlyData) string {
	if len(hourly) == 0 {
		return ""
	}
	type kv struct {
		hour string
		max  float64
	}
	var arr []kv
	for h, d := range hourly {
		arr = append(arr, kv{h, d.MaxPower})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].max == arr[j].max {
			return arr[i].hour < arr[j].hour
		}
		return arr[i].max > arr[j].max
	})
	return arr[0].hour // "HH"
}

func averageFloat(get func(i int) float64, n int) float64 {
	if n == 0 {
		return 0
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += get(i)
	}
	return sum / float64(n)
}

func stddevFloat(get func(i int) float64, n int, mean float64) float64 {
	if n == 0 {
		return 0
	}
	var v float64
	for i := 0; i < n; i++ {
		d := get(i) - mean
		v += d * d
	}
	return math.Sqrt(v / float64(n))
}

func round2(x float64) float64 { return math.Round(x*100) / 100 }
func round3(x float64) float64 { return math.Round(x*1000) / 1000 }
func roundSlice(xs []float64, places int) []float64 {
	if xs == nil {
		return nil
	}
	out := make([]float64, len(xs))
	k := math.Pow10(places)
	for i, v := range xs {
		out[i] = math.Round(v*k) / k
	}
	return out
}
func max0(x float64) float64 {
	if x < 0 {
		return 0
	}
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0
	}
	return x
}

// --- Persistence & reporting ---

func storeAnalyticsSummary(ctx context.Context, facilityID string, analytics DailyAnalytics) error {
	if facilityID == "" {
		return errors.New("facilityID is empty")
	}

	// Flatten for DDB: include facilityId + date as the composite key if your table expects it.
	item := map[string]interface{}{
		"facilityId":          facilityID,
		"date":                analytics.Date,
		"readingCount":        analytics.ReadingCount,
		"totalConsumption":    analytics.TotalConsumption,
		"totalConsumptionMWh": analytics.TotalConsumptionMWh,
		"averagePower":        analytics.AveragePower,
		"peakPower":           analytics.PeakPower,
		"minPower":            analytics.MinPower,
		"avgVoltage":          analytics.AvgVoltage,
		"voltageStdDev":       analytics.VoltageStdDev,
		"avgCurrent":          analytics.AvgCurrent,
		"powerFactor":         analytics.PowerFactor,
		"peakHour":            analytics.PeakHour,
		"hourlyData":          analytics.HourlyData,
		"createdAt":           analytics.CreatedAt,
	}

	marshalled, err := ddbattr.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal analytics: %w", err)
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableAnalytics),
		Item:      marshalled,
	})
	if err != nil {
		return fmt.Errorf("put item: %w", err)
	}
	return nil
}

func generateReport(ctx context.Context, facilityID, date string, analytics DailyAnalytics) (string, error) {
	report := map[string]interface{}{
		"title":       fmt.Sprintf("Daily Energy Report - %s", facilityID),
		"date":        date,
		"generatedAt": time.Now().Format(time.RFC3339),
		"summary": map[string]interface{}{
			"total_consumption": fmt.Sprintf("%.2f kWh", analytics.TotalConsumption),
			"average_power":     fmt.Sprintf("%.2f kW", analytics.AveragePower),
			"peak_power":        fmt.Sprintf("%.2f kW", analytics.PeakPower),
			"peak_hour":         fmt.Sprintf("%s:00", analytics.PeakHour),
			"power_factor":      analytics.PowerFactor,
			"reading_count":     analytics.ReadingCount,
		},
		"hourly_breakdown": analytics.HourlyData,
		"recommendations":  generateRecommendations(analytics),
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report: %w", err)
	}

	key := fmt.Sprintf("reports/%s/%s-analytics.json", safePath(facilityID), date)
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(b),
		ContentType: aws.String("application/json"),
		Metadata: map[string]string{
			"facility-id":  facilityID,
			"report-date":  date,
			"generated-at": time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("s3 put: %w", err)
	}

	// Virtual-hosted–style URL (region-agnostic for public buckets or signed-URL flows)
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3Bucket, url.PathEscape(key)), nil
}

func safePath(s string) string {
	// simple sanitizer for S3 key path components
	return url.PathEscape(s)
}

// --- Recommendations ---

func generateRecommendations(a DailyAnalytics) []map[string]string {
	var recs []map[string]string

	if a.AveragePower > 50 {
		recs = append(recs, map[string]string{
			"priority": "high",
			"category": "consumption",
			"message":  "Average power is high. Consider load shifting and efficiency measures.",
		})
	}

	if a.PowerFactor < 0.85 && a.PowerFactor > 0 {
		recs = append(recs, map[string]string{
			"priority": "medium",
			"category": "efficiency",
			"message":  fmt.Sprintf("Low power factor (%.3f). Evaluate correction equipment.", a.PowerFactor),
		})
	}

	if a.VoltageStdDev > 10 {
		recs = append(recs, map[string]string{
			"priority": "high",
			"category": "quality",
			"message":  "High voltage variability detected. Inspect electrical infrastructure.",
		})
	}

	if a.PeakHour != "" {
		if h, _ := strconv.Atoi(a.PeakHour); h >= 9 && h <= 17 {
			recs = append(recs, map[string]string{
				"priority": "low",
				"category": "optimization",
				"message":  fmt.Sprintf("Peak at %s:00. Shift non-critical loads to off-peak hours.", a.PeakHour),
			})
		}
	}

	return recs
}

func main() {
	lambda.Start(Handler)
}
