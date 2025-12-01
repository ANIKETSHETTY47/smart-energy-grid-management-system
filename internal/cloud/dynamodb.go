package cloud

import (
	"context"
	"fmt"
	"strconv"
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
func NewDynamoDBClient(region string) (*DynamoDBClient, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &DynamoDBClient{
		svc: dynamodb.NewFromConfig(cfg),
		ctx: ctx,
	}, nil
}

// Reading represents the DynamoDB structure for energy readings (actual structure in DB)
type Reading struct {
	DeviceID    string  `dynamodbav:"device_id"`
	Timestamp   int64   `dynamodbav:"timestamp"`
	Voltage     float64 `dynamodbav:"voltage"`
	Current     float64 `dynamodbav:"current"`
	PowerKW     float64 `dynamodbav:"power_kw"`
	Temperature float64 `dynamodbav:"temperature,omitempty"`
	Status      string  `dynamodbav:"status,omitempty"`
}

// PutReading stores an energy reading in DynamoDB
func (c *DynamoDBClient) PutReading(reading *domain.Reading, facilityID string) error {
	deviceID := fmt.Sprintf("%s-meter-%d", facilityID, reading.MeterID)
	dbReading := Reading{
		DeviceID:  deviceID,
		Timestamp: reading.Timestamp.Unix(),
		Voltage:   reading.Voltage,
		Current:   reading.Current,
		PowerKW:   reading.PowerKW,
		Status:    "operational",
	}

	item, err := attributevalue.MarshalMap(dbReading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("energy-grid-readings"),
		Item:      item,
	}

	_, err = c.svc.PutItem(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	return nil
}

// GetRecentReadings retrieves recent readings for a facility (supports ANY device_id format)
// GetRecentReadings retrieves all readings from DynamoDB within the specified duration.
// It scans the table for entries with timestamps greater than (now - duration).
// Device IDs are parsed to extract meter IDs, supporting formats like "device-001" and "facility-001-meter-1".
// Returns a slice of domain.Reading or an error if the scan or unmarshal fails.
func (c *DynamoDBClient) GetRecentReadings(facilityID string, duration time.Duration) ([]domain.Reading, error) {
	startTime := time.Now().Add(-duration).Unix()

	// Scan all items with timestamp filter (works with device-001, device-002, etc.)
	input := &dynamodb.ScanInput{
		TableName:        aws.String("energy-grid-readings"),
		FilterExpression: aws.String("#ts > :startTime"),
		ExpressionAttributeNames: map[string]string{
			"#ts": "timestamp",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":startTime": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", startTime)},
		},
	}

	result, err := c.svc.Scan(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB: %w", err)
	}

	var dbReadings []Reading
	err = attributevalue.UnmarshalListOfMaps(result.Items, &dbReadings)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal readings: %w", err)
	}

	// Convert to domain.Reading format
	readings := make([]domain.Reading, 0, len(dbReadings))
	for _, r := range dbReadings {
		// Extract meter ID from device_id (supports: device-001, device-002, facility-001-meter-1, etc.)
		meterID := int64(1)

		// Try facility-meter format first
		_, err := fmt.Sscanf(r.DeviceID, facilityID+"-meter-%d", &meterID)
		if err != nil {
			// Try device-XXX format
			_, err = fmt.Sscanf(r.DeviceID, "device-%d", &meterID)
		}

		readings = append(readings, domain.Reading{
			MeterID:   meterID,
			Timestamp: time.Unix(r.Timestamp, 0),
			Voltage:   r.Voltage,
			Current:   r.Current,
			PowerKW:   r.PowerKW,
		})
	}

	return readings, nil
}

// Alert represents an alert stored in DynamoDB
type Alert struct {
	AlertID      string                 `dynamodbav:"alert_id" json:"alert_id"`
	CreatedAt    int64                  `dynamodbav:"created_at" json:"created_at"`
	FacilityID   string                 `dynamodbav:"facility_id,omitempty" json:"facility_id"`
	Severity     string                 `dynamodbav:"severity" json:"severity"`
	Type         string                 `dynamodbav:"type" json:"type"`
	Message      string                 `dynamodbav:"message" json:"message"`
	Acknowledged bool                   `dynamodbav:"acknowledged" json:"acknowledged"`
	EquipmentID  string                 `dynamodbav:"equipment_id" json:"equipment_id"`
	Metadata     map[string]interface{} `dynamodbav:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp    int64                  `json:"timestamp"` // For frontend compatibility
}

// CreateAlert stores a new alert in DynamoDB
func (c *DynamoDBClient) CreateAlert(facilityID, equipmentID, severity, alertType, message string) error {
	alert := Alert{
		AlertID:      fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		CreatedAt:    time.Now().Unix(),
		FacilityID:   facilityID,
		Severity:     severity,
		Type:         alertType,
		Message:      message,
		Acknowledged: false,
		EquipmentID:  equipmentID,
		Timestamp:    time.Now().Unix(),
	}

	item, err := attributevalue.MarshalMap(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("energy-grid-alerts"),
		Item:      item,
	}

	_, err = c.svc.PutItem(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetAlerts retrieves alerts (works without facility_id filter)
func (c *DynamoDBClient) GetAlerts(facilityID string, severityFilter *string) ([]Alert, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String("energy-grid-alerts"),
	}

	// Add severity filter if provided
	if severityFilter != nil && *severityFilter != "" {
		input.FilterExpression = aws.String("severity = :sev")
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":sev": &types.AttributeValueMemberS{Value: *severityFilter},
		}
	}

	result, err := c.svc.Scan(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan alerts: %w", err)
	}

	var alerts []Alert
	err = attributevalue.UnmarshalListOfMaps(result.Items, &alerts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal alerts: %w", err)
	}

	// Set Timestamp field for frontend
	for i := range alerts {
		alerts[i].Timestamp = alerts[i].CreatedAt
		if alerts[i].FacilityID == "" {
			alerts[i].FacilityID = "facility-001"
		}
	}

	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (c *DynamoDBClient) AcknowledgeAlert(alertID string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("energy-grid-alerts"),
		Key: map[string]types.AttributeValue{
			"alert_id": &types.AttributeValueMemberS{Value: alertID},
		},
		UpdateExpression: aws.String("SET acknowledged = :ack"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ack": &types.AttributeValueMemberBOOL{Value: true},
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
	EquipmentID     string  `dynamodbav:"equipment_id" json:"id"`
	FacilityID      string  `dynamodbav:"facility_id,omitempty" json:"facility_id"`
	Type            string  `dynamodbav:"type" json:"type"`
	Status          string  `dynamodbav:"status" json:"status"`
	InstallDate     int64   `dynamodbav:"install_date" json:"install_date"`
	LastMaintenance int64   `dynamodbav:"last_maintenance" json:"last_maintenance"`
	HealthScore     float64 `dynamodbav:"health_score" json:"health"`
}

// GetEquipment retrieves all equipment
func (c *DynamoDBClient) GetEquipment(facilityID string) ([]Equipment, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String("energy-grid-equipment"),
	}

	result, err := c.svc.Scan(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan equipment: %w", err)
	}

	var equipment []Equipment
	err = attributevalue.UnmarshalListOfMaps(result.Items, &equipment)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
	}

	// Set facility_id if missing
	for i := range equipment {
		if equipment[i].FacilityID == "" {
			equipment[i].FacilityID = "facility-001"
		}
	}

	return equipment, nil
}

// UpdateEquipmentHealth updates the health score of equipment
func (c *DynamoDBClient) UpdateEquipmentHealth(equipmentID string, healthScore float64) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("energy-grid-equipment"),
		Key: map[string]types.AttributeValue{
			"equipment_id": &types.AttributeValueMemberS{Value: equipmentID},
		},
		UpdateExpression: aws.String("SET health_score = :score, last_checked = :time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":score": &types.AttributeValueMemberN{Value: strconv.FormatFloat(healthScore, 'f', 2, 64)},
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
func (c *DynamoDBClient) BatchPutReadings(readings []domain.Reading, facilityID string) error {
	const batchSize = 25

	for i := 0; i < len(readings); i += batchSize {
		end := i + batchSize
		if end > len(readings) {
			end = len(readings)
		}

		batch := readings[i:end]
		writeRequests := make([]types.WriteRequest, len(batch))

		for j, reading := range batch {
			deviceID := fmt.Sprintf("%s-meter-%d", facilityID, reading.MeterID)
			dbReading := Reading{
				DeviceID:  deviceID,
				Timestamp: reading.Timestamp.Unix(),
				Voltage:   reading.Voltage,
				Current:   reading.Current,
				PowerKW:   reading.PowerKW,
				Status:    "operational",
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
				"energy-grid-readings": writeRequests,
			},
		}

		_, err := c.svc.BatchWriteItem(c.ctx, input)
		if err != nil {
			return fmt.Errorf("failed to batch write items: %w", err)
		}
	}

	return nil
}
