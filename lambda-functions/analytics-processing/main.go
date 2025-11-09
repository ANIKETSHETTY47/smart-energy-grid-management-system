package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/aggregator"
	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/converter"
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
	VoltageVariance     float64               `json:"voltage_variance"`
	AvgCurrent          float64               `json:"avg_current"`
	PowerFactor         float64               `json:"power_factor"`
	PeakHour            string                `json:"peak_hour"`
	HourlyData          map[string]HourlyData `json:"hourly_data"`
	CreatedAt           int64                 `dynamodbav:"createdAt" json:"created_at"`
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
	// YOUR ORIGINAL CONTRIBUTION: Convert to library points
	points := make([]aggregator.Point, len(readings))
	for i, r := range readings {
		points[i] = aggregator.Point{
			Value:     r.PowerKW,
			Timestamp: time.Unix(r.Timestamp, 0),
		}
	}

	// YOUR ORIGINAL CONTRIBUTION: Use library aggregation
	totalPower := aggregator.Sum(points)
	avgPower := aggregator.Average(points)
	// YOUR ORIGINAL CONTRIBUTION: Calculate moving average
	movingAvg := aggregator.MovingAverage(points, 12) // 12-point moving average

	// YOUR ORIGINAL CONTRIBUTION: Use converter for cost calculations
	conv := &converter.EnergyConverter{}
	totalConsumptionMWh := conv.KWhToMWh(totalPower)

	// Calculate costs for different tiers
	peakCost := conv.CalculateCost(totalPower*0.4, 0.20, "peak")       // 40% peak hours
	offPeakCost := conv.CalculateCost(totalPower*0.6, 0.20, "offpeak") // 60% off-peak
	totalCost := peakCost + offPeakCost

	analytics := DailyAnalytics{
		Date:                date,
		ReadingCount:        len(readings),
		TotalConsumption:    totalPower,
		TotalConsumptionMWh: totalConsumptionMWh,
		AveragePower:        avgPower,
		PeakPower:           findMax(points),
		MinPower:            findMin(points),
		MovingAverage:       movingAvg,
		EstimatedCost:       totalCost,
		CostBreakdown: map[string]float64{
			"peak":    peakCost,
			"offpeak": offPeakCost,
		},
		HourlyData: calculateHourlyData(readings),
		CreatedAt:  time.Now().Unix(),
	}

	// Calculate additional metrics
	analytics.AvgVoltage = calculateAvgVoltage(readings)
	analytics.AvgCurrent = calculateAvgCurrent(readings)
	analytics.VoltageVariance = calculateVoltageVariance(readings, analytics.AvgVoltage)

	// YOUR ORIGINAL CONTRIBUTION: Calculate efficiency using library
	apparentPower := analytics.AvgVoltage * analytics.AvgCurrent
	analytics.PowerFactor = conv.CalculateEfficiency(apparentPower, avgPower)

	return analytics
}

func findMax(points []aggregator.Point) float64 {
	max := 0.0
	for _, p := range points {
		if p.Value > max {
			max = p.Value
		}
	}
	return max
}

func findMin(points []aggregator.Point) float64 {
	min := math.MaxFloat64
	for _, p := range points {
		if p.Value < min {
			min = p.Value
		}
	}
	return min
}

func calculateHourlyData(readings []Reading) map[string]HourlyData {
	hourlyMap := make(map[string]HourlyData)

	for _, r := range readings {
		hour := time.Unix(r.Timestamp, 0).Format("15")
		data := hourlyMap[hour]
		data.Count++
		data.TotalPower += r.PowerKW
		if r.PowerKW > data.MaxPower {
			data.MaxPower = r.PowerKW
		}
		hourlyMap[hour] = data
	}

	// Calculate averages
	for hour, data := range hourlyMap {
		if data.Count > 0 {
			data.AvgPower = data.TotalPower / float64(data.Count)
			hourlyMap[hour] = data
		}
	}

	return hourlyMap
}

func calculateAvgVoltage(readings []Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range readings {
		sum += r.Voltage
	}
	return sum / float64(len(readings))
}

func calculateAvgCurrent(readings []Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range readings {
		sum += r.Current
	}
	return sum / float64(len(readings))
}

func calculateVoltageVariance(readings []Reading, avgVoltage float64) float64 {
	if len(readings) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range readings {
		diff := r.Voltage - avgVoltage
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(readings)))
}

// YOUR ORIGINAL CONTRIBUTION: Store analytics summary in DynamoDB
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
