package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/cloud"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/config"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/domain"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/repository"
	"github.com/jmoiron/sqlx"
)

type Services struct {
	*repository.Repos
	Readings  *ReadingService
	Analytics *AnalyticsService
	Alerts    *AlertService

	// Cloud clients
	DynamoDB *cloud.DynamoDBClient
	S3       *cloud.S3Client
	SNS      *cloud.SNSClient
	UseCloud bool
}

// New creates a new Services instance with cloud integration
func New(db *sqlx.DB) (*Services, error) {
	repos := repository.New(db)

	svcs := &Services{
		Repos:    repos,
		UseCloud: config.UseCloudServices(),
	}

	// Initialize cloud clients if enabled
	if svcs.UseCloud {
		var err error

		svcs.DynamoDB, err = cloud.NewDynamoDBClient(config.AWSRegion())
		if err != nil {
			return nil, fmt.Errorf("failed to init DynamoDB: %w", err)
		}

		svcs.S3, err = cloud.NewS3Client(config.AWSRegion(), config.S3Bucket())
		if err != nil {
			return nil, fmt.Errorf("failed to init S3: %w", err)
		}

		svcs.SNS, err = cloud.NewSNSClient(config.AWSRegion(), config.SNSTopicArn())
		if err != nil {
			return nil, fmt.Errorf("failed to init SNS: %w", err)
		}
	}

	svcs.Readings = &ReadingService{
		repos:    repos,
		dynamoDB: svcs.DynamoDB,
		useCloud: svcs.UseCloud,
	}

	svcs.Analytics = &AnalyticsService{
		repos:    repos,
		dynamoDB: svcs.DynamoDB,
		useCloud: svcs.UseCloud,
	}

	svcs.Alerts = &AlertService{
		repos:    repos,
		dynamoDB: svcs.DynamoDB,
		sns:      svcs.SNS,
		useCloud: svcs.UseCloud,
	}

	return svcs, nil
}

type ReadingService struct {
	repos    *repository.Repos
	dynamoDB *cloud.DynamoDBClient
	useCloud bool
}

// FromMQTT processes MQTT message and stores in appropriate backend
func (s *ReadingService) FromMQTT(topic string, payload []byte) error {
	var r struct {
		MeterID   string    `json:"meter_id"`
		Timestamp time.Time `json:"timestamp"`
		Voltage   float64   `json:"voltage"`
		Current   float64   `json:"current"`
		PowerKW   float64   `json:"power_kw"`
	}
	if err := json.Unmarshal(payload, &r); err != nil {
		return err
	}

	rd := &domain.Reading{
		MeterID:   1, // demo mapping
		Timestamp: r.Timestamp,
		Voltage:   r.Voltage,
		Current:   r.Current,
		PowerKW:   r.PowerKW,
	}

	// Store in cloud if enabled, otherwise use local DB
	if s.useCloud {
		return s.dynamoDB.PutReading(rd, "facility-001")
	}

	return s.repos.InsertReading(rd)
}
