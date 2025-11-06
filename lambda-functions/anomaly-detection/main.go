package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// YOUR ORIGINAL CONTRIBUTION: Lambda handler for real-time anomaly detection
// Analyzes incoming energy readings and triggers alerts for abnormal consumption

var (
	dynamoClient *dynamodb.Client
	snsClient    *sns.Client
	topicArn     string
	ctx          = context.Background()
)

// Reading represents an energy reading from DynamoDB
type Reading struct {
	FacilityID  string  `dynamodbav:"facilityId" json:"facility_id"`
	MeterID     string  `dynamodbav:"meterId" json:"meter_id"`
	Timestamp   int64   `dynamodbav:"timestamp" json:"timestamp"`
	Voltage     float64 `dynamodbav:"voltage" json:"voltage"`
	Current     float64 `dynamodbav:"current" json:"current"`
	PowerKW     float64 `dynamodbav:"powerKw" json:"power_kw"`
	Status      string  `dynamodbav:"status" json:"status"`
	Temperature float64 `dynamodbav:"temperature" json:"temperature"`
}

// Alert represents an alert to be stored
type Alert struct {
	AlertID      string                 `dynamodbav:"alertId"`
	FacilityID   string                 `dynamodbav:"facilityId"`
	Timestamp    int64                  `dynamodbav:"timestamp"`
	Severity     string                 `dynamodbav:"severity"`
	Type         string                 `dynamodbav:"type"`
	Message      string                 `dynamodbav:"message"`
	Acknowledged bool                   `dynamodbav:"acknowledged"`
	EquipmentID  string                 `dynamodbav:"equipmentId"`
	Metadata     map[string]interface{} `dynamodbav:"metadata"`
}

// AnomalyResult holds anomaly detection results
type AnomalyResult struct {
	IsAnomaly        bool    `json:"is_anomaly"`
	CurrentPower     float64 `json:"current_power"`
	Mean             float64 `json:"mean"`
	StdDev           float64 `json:"std_dev"`
	Threshold        float64 `json:"threshold"`
	DeviationPercent float64 `json:"deviation_percent"`
	Severity         string  `json:"severity"`
	Reason           string  `json:"reason"`
}

func init() {
	// YOUR ORIGINAL CONTRIBUTION: Initialize AWS clients with SDK v2
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	snsClient = sns.NewFromConfig(cfg)
	topicArn = os.Getenv("SNS_TOPIC_ARN")
}

// Handler is the Lambda entry point
func Handler(ctx context.Context, event events.DynamoDBEvent) error {
	fmt.Printf("Processing %d records\n", len(event.Records))

	for _, record := range event.Records {
		if record.EventName != "INSERT" && record.EventName != "MODIFY" {
			continue
		}

		// YOUR ORIGINAL CONTRIBUTION: Parse DynamoDB stream event
		reading, err := parseReading(record.Change.NewImage)
		if err != nil {
			fmt.Printf("Error parsing reading: %v\n", err)
			continue
		}

		fmt.Printf("Processing reading for facility %s, meter %s\n", reading.FacilityID, reading.MeterID)

		// Get historical readings for comparison
		historical, err := getHistoricalReadings(reading.FacilityID, reading.MeterID, 24)
		if err != nil {
			fmt.Printf("Error fetching historical readings: %v\n", err)
			continue
		}

		// Detect anomaly
		anomaly := detectAnomaly(reading, historical)

		if anomaly.IsAnomaly {
			fmt.Printf("Anomaly detected: %+v\n", anomaly)

			// Store alert in DynamoDB
			if err := storeAlert(reading, anomaly); err != nil {
				fmt.Printf("Error storing alert: %v\n", err)
			}

			// Send SNS notification
			if err := sendAlert(reading, anomaly); err != nil {
				fmt.Printf("Error sending SNS notification: %v\n", err)
			}
		}
	}

	return nil
}

// YOUR ORIGINAL CONTRIBUTION: Parse DynamoDB AttributeValue map to Reading struct

func parseReading(image map[string]events.DynamoDBAttributeValue) (*Reading, error) {
	reading := &Reading{}

	if v, ok := image["facilityId"]; ok {
		reading.FacilityID = v.String()
	}
	if v, ok := image["meterId"]; ok {
		reading.MeterID = v.String()
	}
	if v, ok := image["timestamp"]; ok {
		if ts, err := strconv.ParseInt(v.Number(), 10, 64); err == nil {
			reading.Timestamp = ts
		}
	}
	if v, ok := image["voltage"]; ok {
		if val, err := strconv.ParseFloat(v.Number(), 64); err == nil {
			reading.Voltage = val
		}
	}
	if v, ok := image["current"]; ok {
		if val, err := strconv.ParseFloat(v.Number(), 64); err == nil {
			reading.Current = val
		}
	}
	if v, ok := image["powerKw"]; ok {
		if val, err := strconv.ParseFloat(v.Number(), 64); err == nil {
			reading.PowerKW = val
		}
	}

	return reading, nil
}

// YOUR ORIGINAL CONTRIBUTION: Retrieve historical readings for baseline calculation
func getHistoricalReadings(facilityID, meterID string, hours int) ([]Reading, error) {
	now := time.Now().Unix()
	startTime := now - int64(hours*3600)

	input := &dynamodb.QueryInput{
		TableName:              aws.String("EnergyReadings"),
		KeyConditionExpression: aws.String("facilityId = :fid AND #ts BETWEEN :start AND :end"),
		ExpressionAttributeNames: map[string]string{
			"#ts": "timestamp",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fid":   &types.AttributeValueMemberS{Value: facilityID},
			":start": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", startTime)},
			":end":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
		},
		Limit: aws.Int32(100),
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

// YOUR ORIGINAL CONTRIBUTION: Statistical anomaly detection using mean and standard deviation
func detectAnomaly(current *Reading, historical []Reading) AnomalyResult {
	if len(historical) < 3 {
		return AnomalyResult{
			IsAnomaly: false,
			Reason:    "Insufficient historical data",
		}
	}

	// Calculate statistics
	var sum float64
	powers := make([]float64, len(historical))
	for i, r := range historical {
		powers[i] = r.PowerKW
		sum += r.PowerKW
	}

	mean := sum / float64(len(historical))

	// Calculate standard deviation
	var varianceSum float64
	for _, p := range powers {
		varianceSum += math.Pow(p-mean, 2)
	}
	variance := varianceSum / float64(len(historical))
	stdDev := math.Sqrt(variance)

	// Check if anomaly
	threshold := mean + (stdDev * 2) // 2 standard deviations
	deviation := math.Abs(current.PowerKW - mean)
	deviationPercent := (deviation / mean) * 100

	thresholdMultiplier := 1.5 // 50% above average
	isAnomaly := current.PowerKW > threshold || current.PowerKW > (mean*thresholdMultiplier)

	severity := "low"
	if deviationPercent > 100 {
		severity = "critical"
	} else if deviationPercent > 50 {
		severity = "high"
	} else if deviationPercent > 25 {
		severity = "medium"
	}

	reason := "Normal"
	if isAnomaly {
		reason = fmt.Sprintf("Power consumption %.1f%% above average", deviationPercent)
	}

	return AnomalyResult{
		IsAnomaly:        isAnomaly,
		CurrentPower:     current.PowerKW,
		Mean:             mean,
		StdDev:           stdDev,
		Threshold:        threshold,
		DeviationPercent: deviationPercent,
		Severity:         severity,
		Reason:           reason,
	}
}

// YOUR ORIGINAL CONTRIBUTION: Store alert in DynamoDB for tracking
func storeAlert(reading *Reading, anomaly AnomalyResult) error {
	alertID := fmt.Sprintf("alert-%d-%d", time.Now().Unix(), time.Now().Nanosecond())

	alert := Alert{
		AlertID:      alertID,
		FacilityID:   reading.FacilityID,
		EquipmentID:  reading.MeterID,
		Timestamp:    time.Now().Unix(),
		Severity:     anomaly.Severity,
		Type:         "anomaly",
		Message:      fmt.Sprintf("Abnormal power consumption: %.2f kW (%.1f%% above average)", anomaly.CurrentPower, anomaly.DeviationPercent),
		Acknowledged: false,
		Metadata: map[string]interface{}{
			"current_power":     anomaly.CurrentPower,
			"average_power":     anomaly.Mean,
			"deviation_percent": anomaly.DeviationPercent,
		},
	}

	item, err := attributevalue.MarshalMap(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Alerts"),
		Item:      item,
	}

	_, err = dynamoClient.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// YOUR ORIGINAL CONTRIBUTION: Send real-time alert via SNS
func sendAlert(reading *Reading, anomaly AnomalyResult) error {
	if topicArn == "" {
		fmt.Println("SNS_TOPIC_ARN not configured, skipping notification")
		return nil
	}

	subject := fmt.Sprintf("[%s] Energy Grid Anomaly - %s", anomaly.Severity, reading.FacilityID)
	message := fmt.Sprintf(`
Energy Grid Anomaly Detected

Facility: %s
Meter: %s
Severity: %s

Current Power: %.2f kW
Average Power: %.2f kW
Deviation: %.1f%%

Threshold: %.2f kW
Time: %s

Reason: %s

Action Required: Please investigate immediately.
`,
		reading.FacilityID,
		reading.MeterID,
		anomaly.Severity,
		anomaly.CurrentPower,
		anomaly.Mean,
		anomaly.DeviationPercent,
		anomaly.Threshold,
		time.Now().Format(time.RFC3339),
		anomaly.Reason,
	)

	input := &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Subject:  aws.String(subject),
		Message:  aws.String(message),
	}

	result, err := snsClient.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish to SNS: %w", err)
	}

	fmt.Printf("SNS notification sent: %s\n", *result.MessageId)
	return nil
}

func main() {
	lambda.Start(Handler)
}
