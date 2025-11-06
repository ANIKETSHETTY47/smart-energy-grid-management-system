package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// YOUR ORIGINAL CONTRIBUTION: Lambda function for daily analytics aggregation
// Processes energy readings and generates daily summaries

var (
	dynamoClient *dynamodb.Client
	s3Client     *s3.Client
	s3Bucket     string
	ctx          = context.Background()
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
	Date             string                `json:"date"`
	ReadingCount     int                   `json:"reading_count"`
	TotalConsumption float64               `json:"total_consumption"`
	AveragePower     float64               `json:"average_power"`
	PeakPower        float64               `json:"peak_power"`
	MinPower         float64               `json:"min_power"`
	AvgVoltage       float64               `json:"avg_voltage"`
	VoltageVariance  float64               `json:"voltage_variance"`
	AvgCurrent       float64               `json:"avg_current"`
	PowerFactor      float64               `json:"power_factor"`
	PeakHour         string                `json:"peak_hour"`
	HourlyData       map[string]HourlyData `json:"hourly_data"`
	CreatedAt        int64                 `dynamodbav:"createdAt" json:"created_at"`
}

type LambdaEvent struct {
	Date       string `json:"date"`
	FacilityID string `json:"facility_id"`
}

type LambdaResponse struct {
	StatusCode int                    `json:"statusCode"`
	Body       map[string]interface{} `json:"body"`
}

func init() {
	// YOUR ORIGINAL CONTRIBUTION: Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
	s3Bucket = os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "energy-grid-reports"
	}
}

func Handler(ctx context.Context, event LambdaEvent) (LambdaResponse, error) {
	fmt.Printf("Analytics processing started for date: %s, facility: %s\n", event.Date, event.FacilityID)

	// Use yesterday if no date provided
	date := event.Date
	if date == "" {
		yesterday := time.Now().AddDate(0, 0, -1)
		date = yesterday.Format("2006-01-02")
	}

	facilityID := event.FacilityID
	if facilityID == "" {
		facilityID = "facility-001"
	}

	// Fetch readings for the date
	readings, err := getReadingsForDate(facilityID, date)
	if err != nil {
		return LambdaResponse{
			StatusCode: 500,
			Body:       map[string]interface{}{"error": err.Error()},
		}, err
	}

	if len(readings) == 0 {
		return LambdaResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"message": "No data to process"},
		}, nil
	}

	// Calculate analytics
	analytics := calculateDailyAnalytics(readings, date)

	// Store analytics summary in DynamoDB
	if err := storeAnalyticsSummary(facilityID, analytics); err != nil {
		fmt.Printf("Error storing analytics: %v\n", err)
	}

	// Generate and upload report to S3
	reportURL, err := generateReport(facilityID, date, analytics)
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
	}

	return LambdaResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"message":    "Analytics processed successfully",
			"date":       date,
			"analytics":  analytics,
			"report_url": reportURL,
		},
	}, nil
}

// YOUR ORIGINAL CONTRIBUTION: Fetch all readings for a specific date
func getReadingsForDate(facilityID, date string) ([]Reading, error) {
	startOfDay, _ := time.Parse("2006-01-02", date)
	endOfDay := startOfDay.Add(24 * time.Hour)

	startTimestamp := startOfDay.Unix()
	endTimestamp := endOfDay.Unix()

	input := &dynamodb.QueryInput{
		TableName:              aws.String("EnergyReadings"),
		KeyConditionExpression: aws.String("facilityId = :fid AND #ts BETWEEN :start AND :end"),
		ExpressionAttributeNames: map[string]string{
			"#ts": "timestamp",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fid":   &types.AttributeValueMemberS{Value: facilityID},
			":start": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", startTimestamp)},
			":end":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", endTimestamp)},
		},
	}

	result, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	var readings []Reading
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &readings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal readings: %w", err)
	}

	return readings, nil
}

// YOUR ORIGINAL CONTRIBUTION: Calculate comprehensive daily analytics
func calculateDailyAnalytics(readings []Reading, date string) DailyAnalytics {
	analytics := DailyAnalytics{
		Date:         date,
		ReadingCount: len(readings),
		MinPower:     math.MaxFloat64,
		HourlyData:   make(map[string]HourlyData),
		CreatedAt:    time.Now().Unix(),
	}

	var totalPower, totalVoltage, totalCurrent float64

	// First pass: calculate totals and extremes
	for _, r := range readings {
		totalPower += r.PowerKW
		totalVoltage += r.Voltage
		totalCurrent += r.Current

		if r.PowerKW > analytics.PeakPower {
			analytics.PeakPower = r.PowerKW
		}
		if r.PowerKW < analytics.MinPower {
			analytics.MinPower = r.PowerKW
		}

		// Aggregate by hour
		t := time.Unix(r.Timestamp, 0)
		hour := fmt.Sprintf("%02d", t.Hour())

		hourData := analytics.HourlyData[hour]
		hourData.Count++
		hourData.TotalPower += r.PowerKW
		if r.PowerKW > hourData.MaxPower {
			hourData.MaxPower = r.PowerKW
		}
		analytics.HourlyData[hour] = hourData
	}

	// Calculate averages
	count := float64(len(readings))
	analytics.TotalConsumption = totalPower
	analytics.AveragePower = totalPower / count
	analytics.AvgVoltage = totalVoltage / count
	analytics.AvgCurrent = totalCurrent / count

	// Calculate voltage variance (standard deviation)
	var varianceSum float64
	for _, r := range readings {
		varianceSum += math.Pow(r.Voltage-analytics.AvgVoltage, 2)
	}
	analytics.VoltageVariance = math.Sqrt(varianceSum / count)

	// Calculate power factor (simplified)
	apparentPower := analytics.AvgVoltage * analytics.AvgCurrent
	if apparentPower > 0 {
		analytics.PowerFactor = analytics.AveragePower / apparentPower
	}

	// Calculate hourly averages and find peak hour
	maxHourlyPower := 0.0
	for hour, data := range analytics.HourlyData {
		data.AvgPower = data.TotalPower / float64(data.Count)
		analytics.HourlyData[hour] = data

		if data.AvgPower > maxHourlyPower {
			maxHourlyPower = data.AvgPower
			analytics.PeakHour = hour
		}
	}

	// Round values
	analytics.TotalConsumption = math.Round(analytics.TotalConsumption*100) / 100
	analytics.AveragePower = math.Round(analytics.AveragePower*100) / 100
	analytics.PeakPower = math.Round(analytics.PeakPower*100) / 100
	analytics.MinPower = math.Round(analytics.MinPower*100) / 100
	analytics.AvgVoltage = math.Round(analytics.AvgVoltage*100) / 100
	analytics.VoltageVariance = math.Round(analytics.VoltageVariance*100) / 100
	analytics.AvgCurrent = math.Round(analytics.AvgCurrent*100) / 100
	analytics.PowerFactor = math.Round(analytics.PowerFactor*1000) / 1000

	return analytics
}

// YOUR ORIGINAL CONTRIBUTION: Store analytics summary in DynamoDB
func storeAnalyticsSummary(facilityID string, analytics DailyAnalytics) error {
	// Create a map that includes facilityId
	analyticsMap := map[string]interface{}{
		"facilityId":       facilityID,
		"date":             analytics.Date,
		"readingCount":     analytics.ReadingCount,
		"totalConsumption": analytics.TotalConsumption,
		"averagePower":     analytics.AveragePower,
		"peakPower":        analytics.PeakPower,
		"minPower":         analytics.MinPower,
		"avgVoltage":       analytics.AvgVoltage,
		"voltageVariance":  analytics.VoltageVariance,
		"avgCurrent":       analytics.AvgCurrent,
		"powerFactor":      analytics.PowerFactor,
		"peakHour":         analytics.PeakHour,
		"hourlyData":       analytics.HourlyData,
		"createdAt":        analytics.CreatedAt,
	}

	item, err := attributevalue.MarshalMap(analyticsMap)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("AnalyticsSummaries"),
		Item:      item,
	}

	_, err = dynamoClient.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// YOUR ORIGINAL CONTRIBUTION: Generate and upload report to S3
func generateReport(facilityID, date string, analytics DailyAnalytics) (string, error) {
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

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	key := fmt.Sprintf("reports/%s/%s-analytics.json", facilityID, date)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(reportJSON), // <— use io.Reader directly
		ContentType: aws.String("application/json"),
		Metadata: map[string]string{
			"facility-id":  facilityID,
			"report-date":  date,
			"generated-at": time.Now().Format(time.RFC3339),
		},
	}
	_, err = s3Client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	reportURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3Bucket, key)
	return reportURL, nil
}

// YOUR ORIGINAL CONTRIBUTION: Generate efficiency recommendations
func generateRecommendations(analytics DailyAnalytics) []map[string]string {
	recommendations := []map[string]string{}

	// High consumption recommendation
	if analytics.AveragePower > 50 {
		recommendations = append(recommendations, map[string]string{
			"priority": "high",
			"category": "consumption",
			"message":  "Average power consumption is high. Consider load balancing or energy efficiency measures.",
		})
	}

	// Power factor recommendation
	if analytics.PowerFactor < 0.85 {
		recommendations = append(recommendations, map[string]string{
			"priority": "medium",
			"category": "efficiency",
			"message":  fmt.Sprintf("Low power factor detected (%.3f). Install power factor correction equipment.", analytics.PowerFactor),
		})
	}

	// Voltage stability recommendation
	if analytics.VoltageVariance > 10 {
		recommendations = append(recommendations, map[string]string{
			"priority": "high",
			"category": "quality",
			"message":  "High voltage variance detected. Check electrical infrastructure for issues.",
		})
	}

	// Peak hour recommendation
	if analytics.PeakHour != "" {
		peakHourInt := 0
		fmt.Sscanf(analytics.PeakHour, "%d", &peakHourInt)
		if peakHourInt >= 9 && peakHourInt <= 17 {
			recommendations = append(recommendations, map[string]string{
				"priority": "low",
				"category": "optimization",
				"message":  fmt.Sprintf("Peak consumption at %s:00. Consider shifting non-critical loads to off-peak hours.", analytics.PeakHour),
			})
		}
	}

	return recommendations
}

func main() {
	lambda.Start(Handler)
}
