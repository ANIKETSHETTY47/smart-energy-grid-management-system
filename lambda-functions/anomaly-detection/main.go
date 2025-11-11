package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/anomaly"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ddbattr "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var (
	dynamoClient  *dynamodb.Client
	snsClient     *sns.Client
	topicArn      string
	tableReadings string
	tableAlerts   string
	defaultCtx    = context.Background()
)

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

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustAtoi(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

func mustAtof(s string, def float64) float64 {
	if s == "" {
		return def
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return def
}

func init() {
	region := getenv("AWS_REGION", "us-east-1")

	cfg, err := config.LoadDefaultConfig(defaultCtx, config.WithRegion(region))
	if err != nil {
		panic(fmt.Sprintf("unable to load AWS SDK config: %v", err))
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	snsClient = sns.NewFromConfig(cfg)

	topicArn = os.Getenv("SNS_TOPIC_ARN")
	tableReadings = getenv("DDB_TABLE_READINGS", "EnergyReadings")
	tableAlerts = getenv("DDB_TABLE_ALERTS", "Alerts")

	fmt.Printf("Lambda cold start. Region=%s ReadingsTable=%s AlertsTable=%s Topic=%s\n",
		region, tableReadings, tableAlerts, topicArn)
}

// Handler processes DynamoDB Stream events (INSERT/MODIFY on EnergyReadings)
func Handler(ctx context.Context, event events.DynamoDBEvent) error {
	fmt.Printf("Received %d stream records\n", len(event.Records))

	for i, record := range event.Records {
		if record.EventName != "INSERT" && record.EventName != "MODIFY" {
			continue
		}

		reading, err := parseReading(record.Change.NewImage)
		if err != nil {
			fmt.Printf("Record %d: parse error: %v\n", i, err)
			continue
		}
		if reading.FacilityID == "" || reading.MeterID == "" || reading.Timestamp == 0 {
			fmt.Printf("Record %d: missing key fields (facilityId/meterId/timestamp)\n", i)
			continue
		}

		fmt.Printf("Record %d: facility=%s meter=%s ts=%d power=%.3f kW\n",
			i, reading.FacilityID, reading.MeterID, reading.Timestamp, reading.PowerKW)

		// Tunables via env
		hours := mustAtoi(getenv("HISTORICAL_HOURS", "24"), 24)
		window := mustAtoi(getenv("ANOMALY_WINDOW", "24"), 24)
		threshold := mustAtof(getenv("ANOMALY_THRESHOLD_SIGMA", "2.0"), 2.0)
		maxItems := int32(mustAtoi(getenv("HISTORICAL_LIMIT", "200"), 200))

		historical, err := getHistoricalReadings(ctx, reading.FacilityID, reading.MeterID, hours, maxItems)
		if err != nil {
			fmt.Printf("Record %d: error fetching historical readings: %v\n", i, err)
			continue
		}

		an := detectAnomaly(reading, historical, window, threshold)
		if !an.IsAnomaly {
			continue
		}

		fmt.Printf("Record %d: anomaly: %+v\n", i, an)

		if err := storeAlert(ctx, reading, an); err != nil {
			fmt.Printf("Record %d: error storing alert: %v\n", i, err)
		}

		if err := sendAlert(ctx, reading, an); err != nil {
			fmt.Printf("Record %d: error sending SNS: %v\n", i, err)
		}
	}

	return nil
}

// --- Helpers ---

func parseReading(image map[string]events.DynamoDBAttributeValue) (*Reading, error) {
	if image == nil {
		return nil, errors.New("empty image")
	}

	r := &Reading{}

	if v, ok := image["facilityId"]; ok && v.DataType() == events.DataTypeString {
		r.FacilityID = v.String()
	}
	if v, ok := image["meterId"]; ok && v.DataType() == events.DataTypeString {
		r.MeterID = v.String()
	}
	if v, ok := image["timestamp"]; ok && (v.DataType() == events.DataTypeNumber || v.DataType() == events.DataTypeString) {
		// Streams can deliver numbers as strings; handle both
		if ts, err := strconv.ParseInt(v.String(), 10, 64); err == nil {
			r.Timestamp = ts
		}
	}
	if v, ok := image["voltage"]; ok && (v.DataType() == events.DataTypeNumber || v.DataType() == events.DataTypeString) {
		if f, err := strconv.ParseFloat(v.String(), 64); err == nil {
			r.Voltage = f
		}
	}
	if v, ok := image["current"]; ok && (v.DataType() == events.DataTypeNumber || v.DataType() == events.DataTypeString) {
		if f, err := strconv.ParseFloat(v.String(), 64); err == nil {
			r.Current = f
		}
	}
	if v, ok := image["powerKw"]; ok && (v.DataType() == events.DataTypeNumber || v.DataType() == events.DataTypeString) {
		if f, err := strconv.ParseFloat(v.String(), 64); err == nil {
			r.PowerKW = f
		}
	}

	if r.FacilityID == "" || r.MeterID == "" || r.Timestamp == 0 {
		b, _ := json.Marshal(image)
		return nil, fmt.Errorf("missing required keys; image=%s", string(b))
	}
	return r, nil
}

func getHistoricalReadings(ctx context.Context, facilityID, meterID string, hours int, limit int32) ([]Reading, error) {
	now := time.Now().Unix()
	start := now - int64(hours*3600)

	// Partition key is facilityId, sort key is timestamp.
	// If you also key by meterId, you might need a GSI. Adjust KeyCondition accordingly.
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableReadings),
		KeyConditionExpression: aws.String("facilityId = :fid AND #ts BETWEEN :start AND :end"),
		ExpressionAttributeNames: map[string]string{
			"#ts": "timestamp",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fid":   &types.AttributeValueMemberS{Value: facilityID},
			":start": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", start)},
			":end":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
		},
		// Most recent first helps with “latest context” windows
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(limit),
	}

	out, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("dynamodb query failed: %w", err)
	}

	var all []Reading
	if err := ddbattr.UnmarshalListOfMaps(out.Items, &all); err != nil {
		return nil, fmt.Errorf("unmarshal readings failed: %w", err)
	}

	// Optional: filter by meter when the table PK doesn’t include meterId
	if meterID != "" {
		filtered := all[:0]
		for _, r := range all {
			if r.MeterID == meterID {
				filtered = append(filtered, r)
			}
		}
		all = filtered
	}

	// We sorted desc; detector might not care, but stable ascending is nice
	reverseInPlace(all)
	return all, nil
}

func reverseInPlace(a []Reading) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func detectAnomaly(current *Reading, historical []Reading, window int, sigma float64) AnomalyResult {
	if window <= 0 {
		window = 24
	}
	if sigma <= 0 {
		sigma = 2.0
	}

	// Build input to your library
	n := len(historical)
	lib := make([]anomaly.Reading, 0, n+1)
	for _, r := range historical {
		lib = append(lib, anomaly.Reading{
			Consumption: r.PowerKW,
			Timestamp:   r.Timestamp,
		})
	}
	lib = append(lib, anomaly.Reading{
		Consumption: current.PowerKW,
		Timestamp:   current.Timestamp,
	})

	detector := &anomaly.AnomalyDetector{
		Threshold:  sigma,
		WindowSize: window,
	}

	spikes := detector.DetectSpikes(lib)
	outliers := detector.DetectOutliers(lib)

	mean := calculateMean(historical)
	std := calculateStdDev(historical, mean)

	// Safe deviation % when mean == 0
	devPct := 0.0
	if mean != 0 {
		devPct = ((current.PowerKW - mean) / mean) * 100
	}

	isAnomaly := len(spikes) > 0 || len(outliers) > 0
	severity := "low"
	switch {
	case mean > 0 && current.PowerKW >= mean*2.0:
		severity = "critical"
	case mean > 0 && current.PowerKW >= mean*1.5:
		severity = "high"
	}

	threshold := mean + std*sigma
	if math.IsNaN(threshold) || math.IsInf(threshold, 0) {
		threshold = 0
	}

	// If no history, treat large absolute power as low-severity anomaly to avoid silence.
	if len(historical) == 0 && current.PowerKW > 0 {
		isAnomaly = true
		severity = "low"
	}

	return AnomalyResult{
		IsAnomaly:        isAnomaly,
		CurrentPower:     current.PowerKW,
		Mean:             mean,
		StdDev:           std,
		Threshold:        threshold,
		DeviationPercent: devPct,
		Severity:         severity,
		Reason:           fmt.Sprintf("Window=%d sigma=%.2f spikes=%d outliers=%d", window, sigma, len(spikes), len(outliers)),
	}
}

func calculateMean(readings []Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range readings {
		sum += r.PowerKW
	}
	return sum / float64(len(readings))
}

func calculateStdDev(readings []Reading, mean float64) float64 {
	if len(readings) == 0 {
		return 0
	}
	var v float64
	for _, r := range readings {
		d := r.PowerKW - mean
		v += d * d
	}
	return math.Sqrt(v / float64(len(readings)))
}

func storeAlert(ctx context.Context, reading *Reading, an AnomalyResult) error {
	id := fmt.Sprintf("alert-%d-%d", time.Now().Unix(), time.Now().Nanosecond())

	msg := fmt.Sprintf("Abnormal power consumption: %.2f kW (%.1f%% above average)",
		an.CurrentPower, an.DeviationPercent)

	alert := Alert{
		AlertID:      id,
		FacilityID:   reading.FacilityID,
		EquipmentID:  reading.MeterID,
		Timestamp:    time.Now().Unix(),
		Severity:     an.Severity,
		Type:         "anomaly",
		Message:      msg,
		Acknowledged: false,
		Metadata: map[string]interface{}{
			"current_power":     an.CurrentPower,
			"average_power":     an.Mean,
			"std_dev":           an.StdDev,
			"threshold":         an.Threshold,
			"deviation_percent": an.DeviationPercent,
			"reason":            an.Reason,
		},
	}

	item, err := ddbattr.MarshalMap(alert)
	if err != nil {
		return fmt.Errorf("marshal alert failed: %w", err)
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableAlerts),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put alert failed: %w", err)
	}

	return nil
}

func sendAlert(ctx context.Context, reading *Reading, an AnomalyResult) error {
	if topicArn == "" {
		fmt.Println("SNS_TOPIC_ARN not set; skipping notification")
		return nil
	}

	subject := fmt.Sprintf("[%s] Energy Grid Anomaly - %s", an.Severity, reading.FacilityID)
	if len(subject) > 100 {
		subject = subject[:100]
	}

	message := fmt.Sprintf(
		`Energy Grid Anomaly Detected

Facility: %s
Meter: %s
Severity: %s

Current Power: %.2f kW
Average Power: %.2f kW
Deviation: %.1f%%

Threshold: %.2f kW
Time: %s

Reason: %s

Action Required: Please investigate immediately.`,
		reading.FacilityID,
		reading.MeterID,
		an.Severity,
		an.CurrentPower,
		an.Mean,
		an.DeviationPercent,
		an.Threshold,
		time.Now().Format(time.RFC3339),
		an.Reason,
	)

	_, err := snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Subject:  aws.String(subject),
		Message:  aws.String(message),
	})
	if err != nil {
		return fmt.Errorf("sns publish failed: %w", err)
	}
	return nil
}

func main() {
	lambda.Start(Handler)
}
