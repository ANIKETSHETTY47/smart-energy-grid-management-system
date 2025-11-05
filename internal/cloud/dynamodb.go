package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/domain"
)

// DynamoDBClient wraps AWS DynamoDB client for energy grid operations
type DynamoDBClient struct {
	svc *dynamodb.Client
	ctx context.Context
}

// NewDynamoDBClient creates a new DynamoDB client instance
// YOUR ORIGINAL CONTRIBUTION: Initialize DynamoDB client with AWS SDK v2
func NewDynamoDBClient(region string) (*DynamoDBClient, error) {
	ctx := context.Background()

	// Load AWS configuration from environment/credentials
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &DynamoDBClient{
		svc: dynamodb.NewFromConfig(cfg),
		ctx: ctx,
	}, nil
}

// Reading represents the DynamoDB structure for energy readings
type Reading struct {
	FacilityID  string  `dynamodbav:"facilityId"`
	Timestamp   int64   `dynamodbav:"timestamp"`
	MeterID     string  `dynamodbav:"meterId"`
	Voltage     float64 `dynamodbav:"voltage"`
	Current     float64 `dynamodbav:"current"`
	PowerKW     float64 `dynamodbav:"powerKw"`
	Status      string  `dynamodbav:"status"`
	Temperature float64 `dynamodbav:"temperature"`
}

// PutReading stores an energy reading in DynamoDB
// YOUR ORIGINAL CONTRIBUTION: Store reading with proper type conversion and error handling
func (c *DynamoDBClient) PutReading(reading *domain.Reading, facilityID string) error {
	// Convert domain.Reading to DynamoDB Reading structure
	dbReading := Reading{
		FacilityID:  facilityID,
		Timestamp:   reading.Timestamp.Unix(),
		MeterID:     fmt.Sprintf("%d", reading.MeterID),
		Voltage:     reading.Voltage,
		Current:     reading.Current,
		PowerKW:     reading.PowerKW,
		Status:      "operational",
		Temperature: 45.0, // Default value, can be updated based on your domain model
	}

	// Marshal the reading into DynamoDB attribute values
	item, err := attributevalue.MarshalMap(dbReading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading: %w", err)
	}

	// Put item into DynamoDB table
	input := &dynamodb.PutItemInput{
		TableName: aws.String("EnergyReadings"),
		Item:      item,
	}

	_, err = c.svc.PutItem(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	return nil
}

// GetRecentReadings retrieves recent readings for a facility
// YOUR ORIGINAL CONTRIBUTION: Query DynamoDB with time-based filtering
func (c *DynamoDBClient) GetRecentReadings(facilityID string, duration time.Duration) ([]domain.Reading, error) {
	startTime := time.Now().Add(-duration).Unix()

	// Query DynamoDB for readings within time range
	input := &dynamodb.QueryInput{
		TableName:              aws.String("EnergyReadings"),
		KeyConditionExpression: aws.String("facilityId = :fid AND #ts > :startTime"),
		ExpressionAttributeNames: map[string]string{
			"#ts": "timestamp",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fid":       &types.AttributeValueMemberS{Value: facilityID},
			":startTime": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", startTime)},
		},
	}

	result, err := c.svc.Query(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	// Unmarshal results into domain.Reading slice
	var dbReadings []Reading
	err = attributevalue.UnmarshalListOfMaps(result.Items, &dbReadings)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal readings: %w", err)
	}

	// Convert to domain.Reading format
	readings := make([]domain.Reading, len(dbReadings))
	for i, r := range dbReadings {
		meterID := int64(0)
		fmt.Sscanf(r.MeterID, "%d", &meterID)

		readings[i] = domain.Reading{
			MeterID:   meterID,
			Timestamp: time.Unix(r.Timestamp, 0),
			Voltage:   r.Voltage,
			Current:   r.Current,
			PowerKW:   r.PowerKW,
		}
	}

	return readings, nil
}

// Alert represents an alert stored in DynamoDB
type Alert struct {
	AlertID      string `dynamodbav:"alertId"`
	FacilityID   string `dynamodbav:"facilityId"`
	Timestamp    int64  `dynamodbav:"timestamp"`
	Severity     string `dynamodbav:"severity"`
	Type         string `dynamodbav:"type"`
	Message      string `dynamodbav:"message"`
	Acknowledged bool   `dynamodbav:"acknowledged"`
	EquipmentID  string `dynamodbav:"equipmentId"`
}

// CreateAlert stores a new alert in DynamoDB
// YOUR ORIGINAL CONTRIBUTION: Create alert with auto-generated ID
func (c *DynamoDBClient) CreateAlert(facilityID, equipmentID, severity, alertType, message string) error {
	alert := Alert{
		AlertID:      fmt.Sprintf("alert-%d-%d", time.Now().Unix(), time.Now().Nanosecond()),
		FacilityID:   facilityID,
		Timestamp:    time.Now().Unix(),
		Severity:     severity,
		Type:         alertType,
		Message:      message,
		Acknowledged: false,
		EquipmentID:  equipmentID,
	}

	item, err := attributevalue.MarshalMap(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Alerts"),
		Item:      item,
	}

	_, err = c.svc.PutItem(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetAlerts retrieves alerts for a facility
// YOUR ORIGINAL CONTRIBUTION: Query alerts with optional severity filter
func (c *DynamoDBClient) GetAlerts(facilityID string, severityFilter *string) ([]Alert, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String("Alerts"),
		IndexName:              aws.String("facilityId-timestamp-index"),
		KeyConditionExpression: aws.String("facilityId = :fid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fid": &types.AttributeValueMemberS{Value: facilityID},
		},
		ScanIndexForward: aws.Bool(false), // Sort descending (newest first)
	}

	// Add severity filter if provided
	if severityFilter != nil {
		input.FilterExpression = aws.String("severity = :sev")
		input.ExpressionAttributeValues[":sev"] = &types.AttributeValueMemberS{Value: *severityFilter}
	}

	result, err := c.svc.Query(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}

	var alerts []Alert
	err = attributevalue.UnmarshalListOfMaps(result.Items, &alerts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal alerts: %w", err)
	}

	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
// YOUR ORIGINAL CONTRIBUTION: Update alert status with timestamp
func (c *DynamoDBClient) AcknowledgeAlert(alertID string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("Alerts"),
		Key: map[string]types.AttributeValue{
			"alertId": &types.AttributeValueMemberS{Value: alertID},
		},
		UpdateExpression: aws.String("SET acknowledged = :ack, acknowledgedAt = :time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ack":  &types.AttributeValueMemberBOOL{Value: true},
			":time": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		},
	}

	_, err := c.svc.UpdateItem(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	return nil
}

// Equipment represents equipment data in DynamoDB
type Equipment struct {
	EquipmentID     string  `dynamodbav:"equipmentId"`
	FacilityID      string  `dynamodbav:"facilityId"`
	Type            string  `dynamodbav:"type"`
	InstallDate     int64   `dynamodbav:"installDate"`
	LastMaintenance int64   `dynamodbav:"lastMaintenance"`
	HealthScore     float64 `dynamodbav:"healthScore"`
}

// GetEquipment retrieves all equipment for a facility
// YOUR ORIGINAL CONTRIBUTION: Query equipment with GSI
func (c *DynamoDBClient) GetEquipment(facilityID string) ([]Equipment, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String("Equipment"),
		IndexName:              aws.String("facilityId-index"),
		KeyConditionExpression: aws.String("facilityId = :fid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fid": &types.AttributeValueMemberS{Value: facilityID},
		},
	}

	result, err := c.svc.Query(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query equipment: %w", err)
	}

	var equipment []Equipment
	err = attributevalue.UnmarshalListOfMaps(result.Items, &equipment)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
	}

	return equipment, nil
}

// UpdateEquipmentHealth updates the health score of equipment
// YOUR ORIGINAL CONTRIBUTION: Update equipment health with timestamp
func (c *DynamoDBClient) UpdateEquipmentHealth(equipmentID string, healthScore float64) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("Equipment"),
		Key: map[string]types.AttributeValue{
			"equipmentId": &types.AttributeValueMemberS{Value: equipmentID},
		},
		UpdateExpression: aws.String("SET healthScore = :score, lastChecked = :time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":score": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", healthScore)},
			":time":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		},
	}

	_, err := c.svc.UpdateItem(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update equipment health: %w", err)
	}

	return nil
}

// BatchPutReadings stores multiple readings efficiently
// YOUR ORIGINAL CONTRIBUTION: Batch write for performance optimization
func (c *DynamoDBClient) BatchPutReadings(readings []domain.Reading, facilityID string) error {
	const batchSize = 25 // DynamoDB batch write limit

	for i := 0; i < len(readings); i += batchSize {
		end := i + batchSize
		if end > len(readings) {
			end = len(readings)
		}

		batch := readings[i:end]
		writeRequests := make([]types.WriteRequest, len(batch))

		for j, reading := range batch {
			dbReading := Reading{
				FacilityID:  facilityID,
				Timestamp:   reading.Timestamp.Unix(),
				MeterID:     fmt.Sprintf("%d", reading.MeterID),
				Voltage:     reading.Voltage,
				Current:     reading.Current,
				PowerKW:     reading.PowerKW,
				Status:      "operational",
				Temperature: 45.0,
			}

			item, err := attributevalue.MarshalMap(dbReading)
			if err != nil {
				return fmt.Errorf("failed to marshal reading %d: %w", j, err)
			}

			writeRequests[j] = types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: item,
				},
			}
		}

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				"EnergyReadings": writeRequests,
			},
		}

		_, err := c.svc.BatchWriteItem(c.ctx, input)
		if err != nil {
			return fmt.Errorf("failed to batch write items: %w", err)
		}
	}

	return nil
}
