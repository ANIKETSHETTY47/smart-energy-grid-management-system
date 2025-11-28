package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SNSClient wraps AWS SNS client for notification operations
type SNSClient struct {
	svc      *sns.Client
	topicArn string
	ctx      context.Context
}

// NewSNSClient creates a new SNS client instance
// YOUR ORIGINAL CONTRIBUTION: Initialize SNS client for alert notifications
func NewSNSClient(region, topicArn string) (*SNSClient, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &SNSClient{
		svc:      sns.NewFromConfig(cfg),
		topicArn: topicArn,
		ctx:      ctx,
	}, nil
}

// SendAlert sends an alert notification via SNS
// YOUR ORIGINAL CONTRIBUTION: Publish alert messages to SNS topic
func (c *SNSClient) SendAlert(subject, message string) error {
	input := &sns.PublishInput{
		TopicArn: aws.String(c.topicArn),
		Subject:  aws.String(subject),
		Message:  aws.String(message),
	}

	result, err := c.svc.Publish(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish to SNS: %w", err)
	}

	fmt.Printf("Alert sent successfully. MessageID: %s\n", aws.ToString(result.MessageId))
	return nil
}

// SendAnomalyAlert sends a specific alert for detected anomalies
// YOUR ORIGINAL CONTRIBUTION: Format and send anomaly detection alerts
func (c *SNSClient) SendAnomalyAlert(facilityID, meterID string, consumption, deviation float64) error {
	subject := fmt.Sprintf("Energy Grid Alert: Anomaly Detected at %s", facilityID)
	message := fmt.Sprintf(
		"Anomaly Detection Alert\n\n"+
			"Facility: %s\n"+
			"Meter: %s\n"+
			"Consumption: %.2f kW\n"+
			"Deviation: %.2f%%\n"+
			"Time: %s\n\n"+
			"Please investigate immediately.",
		facilityID,
		meterID,
		consumption,
		deviation,
		time.Now().Format(time.RFC3339),
	)

	return c.SendAlert(subject, message)
}

// SendMaintenanceAlert sends a predictive maintenance alert
// YOUR ORIGINAL CONTRIBUTION: Notify about equipment maintenance needs
func (c *SNSClient) SendMaintenanceAlert(equipmentID string, healthScore float64, predictedDate time.Time) error {
	subject := "Predictive Maintenance Alert"
	message := fmt.Sprintf(
		"Equipment Maintenance Required\n\n"+
			"Equipment ID: %s\n"+
			"Current Health Score: %.2f%%\n"+
			"Predicted Maintenance Date: %s\n\n"+
			"Please schedule maintenance to prevent failures.",
		equipmentID,
		healthScore,
		predictedDate.Format("2006-01-02"),
	)

	return c.SendAlert(subject, message)
}

// SendBatchAlerts sends multiple alerts in one notification
// YOUR ORIGINAL CONTRIBUTION: Aggregate multiple alerts for efficiency
func (c *SNSClient) SendBatchAlerts(alerts []string) error {
	if len(alerts) == 0 {
		return nil
	}

	subject := fmt.Sprintf("Energy Grid: %d Alerts", len(alerts))
	message := "Multiple Alerts Detected:\n\n"

	for i, alert := range alerts {
		message += fmt.Sprintf("%d. %s\n", i+1, alert)
	}

	return c.SendAlert(subject, message)
}
