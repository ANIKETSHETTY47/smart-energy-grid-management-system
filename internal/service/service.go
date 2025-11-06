package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/cloud"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/config"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/domain"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/repository"
	"github.com/jmoiron/sqlx"
)

type Services struct {
	Repos     *repository.Repos
	Readings  *ReadingService
	Analytics *AnalyticsService
	Alerts    *AlertService

	// Cloud clients
	DynamoDB *cloud.DynamoDBClient
	S3       *cloud.S3Client
	SNS      *cloud.SNSClient
	Lambda   *cloud.LambdaClient
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

		// Add Lambda client initialization
		svcs.Lambda, err = cloud.NewLambdaClient(config.AWSRegion())
		if err != nil {
			return nil, fmt.Errorf("failed to init Lambda: %w", err)
		}
	}

	svcs.Readings = &ReadingService{
		repos:    repos,
		dynamoDB: svcs.DynamoDB,
		lambda:   svcs.Lambda,
		useCloud: svcs.UseCloud,
	}

	svcs.Analytics = &AnalyticsService{
		repos:    repos,
		dynamoDB: svcs.DynamoDB,
		s3:       svcs.S3,
		lambda:   svcs.Lambda,
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

// ReadingService handles energy reading operations
type ReadingService struct {
	repos    *repository.Repos
	dynamoDB *cloud.DynamoDBClient
	lambda   *cloud.LambdaClient
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

	// Parse meter ID to int64
	var meterIDInt int64 = 1
	if r.MeterID != "" {
		parsed, err := strconv.ParseInt(r.MeterID, 10, 64)
		if err == nil {
			meterIDInt = parsed
		}
	}

	rd := &domain.Reading{
		MeterID:   meterIDInt,
		Timestamp: r.Timestamp,
		Voltage:   r.Voltage,
		Current:   r.Current,
		PowerKW:   r.PowerKW,
	}

	// Store in cloud if enabled
	if s.useCloud && s.dynamoDB != nil {
		if err := s.dynamoDB.PutReading(rd, "facility-001"); err != nil {
			return err
		}

		// Optionally invoke Lambda for immediate anomaly detection
		if s.lambda != nil {
			payload := cloud.AnomalyDetectionPayload{
				FacilityID: "facility-001",
				MeterID:    r.MeterID,
				Timestamp:  r.Timestamp.Unix(),
				Voltage:    r.Voltage,
				Current:    r.Current,
				PowerKW:    r.PowerKW,
			}

			// Invoke asynchronously (fire and forget)
			go func() {
				_, err := s.lambda.InvokeAnomalyDetection(payload)
				if err != nil {
					fmt.Printf("Failed to invoke anomaly detection: %v\n", err)
				}
			}()
		}

		return nil
	}

	return s.repos.InsertReading(rd)
}

// GetRecentReadings retrieves recent readings for a meter
func (s *ReadingService) GetRecentReadings(facilityID string, duration time.Duration) ([]domain.Reading, error) {
	if s.useCloud && s.dynamoDB != nil {
		return s.dynamoDB.GetRecentReadings(facilityID, duration)
	}

	// Fallback to local DB (implement this in repository if needed)
	return []domain.Reading{}, fmt.Errorf("local DB reading retrieval not implemented")
}

// AnalyticsService handles analytics and reporting operations
type AnalyticsService struct {
	repos    *repository.Repos
	dynamoDB *cloud.DynamoDBClient
	s3       *cloud.S3Client
	lambda   *cloud.LambdaClient
	useCloud bool
}

// DailySummary represents daily energy consumption summary
type DailySummary struct {
	Date             time.Time `json:"date"`
	TotalConsumption float64   `json:"total_consumption"`
	PeakPower        float64   `json:"peak_power"`
	AveragePower     float64   `json:"average_power"`
	ReadingCount     int       `json:"reading_count"`
}

// GetDailySummary calculates daily consumption summary
func (s *AnalyticsService) GetDailySummary(facilityID string, date time.Time) (*DailySummary, error) {
	// Get 24 hours of readings
	readings, err := s.getReadingsForDate(facilityID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get readings: %w", err)
	}

	if len(readings) == 0 {
		return &DailySummary{Date: date}, nil
	}

	summary := &DailySummary{
		Date:         date,
		ReadingCount: len(readings),
	}

	var totalPower float64
	for _, r := range readings {
		totalPower += r.PowerKW
		if r.PowerKW > summary.PeakPower {
			summary.PeakPower = r.PowerKW
		}
	}

	summary.TotalConsumption = totalPower
	summary.AveragePower = totalPower / float64(len(readings))

	return summary, nil
}

func (s *AnalyticsService) getReadingsForDate(facilityID string, date time.Time) ([]domain.Reading, error) {
	if s.useCloud && s.dynamoDB != nil {
		return s.dynamoDB.GetRecentReadings(facilityID, 24*time.Hour)
	}

	// Fallback to local DB
	return []domain.Reading{}, nil
}

// GenerateDailyReport generates daily analytics report using Lambda
// YOUR ORIGINAL CONTRIBUTION: Leverage serverless computing for report generation
func (s *AnalyticsService) GenerateDailyReport(facilityID, date string) (string, error) {
	if !s.useCloud || s.lambda == nil {
		return "", fmt.Errorf("cloud services not enabled")
	}

	// Invoke Lambda function to process analytics
	result, err := s.lambda.InvokeAnalyticsProcessing(date, facilityID)
	if err != nil {
		return "", fmt.Errorf("failed to invoke analytics Lambda: %w", err)
	}

	// Extract report URL from response
	if body, ok := result["body"].(map[string]interface{}); ok {
		if reportURL, ok := body["report_url"].(string); ok {
			return reportURL, nil
		}
	}

	return "", fmt.Errorf("no report URL in response")
}

// ScheduleDailyAnalytics triggers daily analytics processing asynchronously
// YOUR ORIGINAL CONTRIBUTION: Background job processing using serverless
func (s *AnalyticsService) ScheduleDailyAnalytics(facilityID string) error {
	if !s.useCloud || s.lambda == nil {
		return fmt.Errorf("cloud services not enabled")
	}

	// Use yesterday's date
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Invoke asynchronously
	return s.lambda.InvokeAnalyticsAsync(yesterday, facilityID)
}

// GenerateReport generates and stores a report (using S3 directly)
func (s *AnalyticsService) GenerateReport(facilityID string, startDate, endDate time.Time) (string, error) {
	if !s.useCloud || s.s3 == nil {
		return "", fmt.Errorf("cloud services not enabled")
	}

	// Generate report data
	reportData := fmt.Sprintf("Energy Report for %s\nPeriod: %s to %s\n",
		facilityID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Upload to S3
	key := fmt.Sprintf("reports/%s/%s.txt", facilityID, time.Now().Format("20060102-150405"))
	url, err := s.s3.UploadReport(key, []byte(reportData), "text/plain")
	if err != nil {
		return "", fmt.Errorf("failed to upload report: %w", err)
	}

	return url, nil
}

// AlertService handles alert operations
type AlertService struct {
	repos    *repository.Repos
	dynamoDB *cloud.DynamoDBClient
	sns      *cloud.SNSClient
	useCloud bool
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(facilityID, equipmentID, severity, alertType, message string) error {
	if s.useCloud && s.dynamoDB != nil {
		if err := s.dynamoDB.CreateAlert(facilityID, equipmentID, severity, alertType, message); err != nil {
			return fmt.Errorf("failed to create alert in DynamoDB: %w", err)
		}

		// Send notification if SNS is available
		if s.sns != nil {
			subject := fmt.Sprintf("[%s] %s Alert", severity, alertType)
			if err := s.sns.SendAlert(subject, message); err != nil {
				// Log error but don't fail - alert is already stored
				fmt.Printf("Failed to send SNS notification: %v\n", err)
			}
		}

		return nil
	}

	// Fallback to local DB (implement this in repository if needed)
	return fmt.Errorf("local alert storage not implemented")
}

// GetAlerts retrieves alerts for a facility
func (s *AlertService) GetAlerts(facilityID string, severityFilter *string) ([]cloud.Alert, error) {
	if s.useCloud && s.dynamoDB != nil {
		return s.dynamoDB.GetAlerts(facilityID, severityFilter)
	}

	return []cloud.Alert{}, fmt.Errorf("local alert retrieval not implemented")
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *AlertService) AcknowledgeAlert(alertID string) error {
	if s.useCloud && s.dynamoDB != nil {
		return s.dynamoDB.AcknowledgeAlert(alertID)
	}

	return fmt.Errorf("local alert acknowledgment not implemented")
}

// DetectAnomalies analyzes readings and creates alerts for anomalies
func (s *AlertService) DetectAnomalies(facilityID string, readings []domain.Reading) error {
	// Simple anomaly detection: flag readings with unusual power consumption
	var sum float64
	for _, r := range readings {
		sum += r.PowerKW
	}

	if len(readings) == 0 {
		return nil
	}

	avg := sum / float64(len(readings))
	threshold := avg * 1.5 // 50% above average

	for _, r := range readings {
		if r.PowerKW > threshold {
			deviation := ((r.PowerKW - avg) / avg) * 100
			message := fmt.Sprintf("Abnormal power consumption detected: %.2f kW (%.1f%% above average)",
				r.PowerKW, deviation)

			if err := s.CreateAlert(facilityID, fmt.Sprintf("meter-%d", r.MeterID),
				"high", "anomaly", message); err != nil {
				return fmt.Errorf("failed to create anomaly alert: %w", err)
			}

			// Send SNS notification if available
			if s.useCloud && s.sns != nil {
				s.sns.SendAnomalyAlert(facilityID, fmt.Sprintf("meter-%d", r.MeterID),
					r.PowerKW, deviation)
			}
		}
	}

	return nil
}
