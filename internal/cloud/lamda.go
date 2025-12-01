package cloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaClient wraps AWS Lambda client
type LambdaClient struct {
	svc *lambda.Client
	ctx context.Context
}

// NewLambdaClient creates a new Lambda client
func NewLambdaClient(region string) (*LambdaClient, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &LambdaClient{
		svc: lambda.NewFromConfig(cfg),
		ctx: ctx,
	}, nil
}

// AnomalyDetectionPayload is input for anomaly detection
type AnomalyDetectionPayload struct {
	FacilityID string  `json:"facility_id"`
	MeterID    string  `json:"meter_id"`
	Timestamp  int64   `json:"timestamp"`
	Voltage    float64 `json:"voltage"`
	Current    float64 `json:"current"`
	PowerKW    float64 `json:"power_kw"`
}

// AnalyticsProcessingPayload is input for analytics processing
type AnalyticsProcessingPayload struct {
	Date       string `json:"date"`
	FacilityID string `json:"facility_id"`
}

// InvokeAnomalyDetection invokes anomaly detection Lambda
func (c *LambdaClient) InvokeAnomalyDetection(payload AnomalyDetectionPayload) (map[string]interface{}, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	input := &lambda.InvokeInput{
		FunctionName: aws.String("anomaly-detection"),
		Payload:      payloadBytes,
	}

	result, err := c.svc.Invoke(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda: %w", err)
	}

	if result.FunctionError != nil {
		return nil, fmt.Errorf("Lambda function error: %s", *result.FunctionError)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(result.Payload, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

// InvokeAnalyticsProcessing invokes analytics Lambda synchronously
func (c *LambdaClient) InvokeAnalyticsProcessing(date, facilityID string) (map[string]interface{}, error) {
	payload := AnalyticsProcessingPayload{
		Date:       date,
		FacilityID: facilityID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String("analytics-processing"),
		Payload:        payloadBytes,
		InvocationType: "RequestResponse",
	}

	result, err := c.svc.Invoke(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda: %w", err)
	}

	if result.FunctionError != nil {
		return nil, fmt.Errorf("Lambda function error: %s", *result.FunctionError)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(result.Payload, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

// InvokeAnalyticsAsync invokes analytics Lambda asynchronously
func (c *LambdaClient) InvokeAnalyticsAsync(date, facilityID string) error {
	payload := AnalyticsProcessingPayload{
		Date:       date,
		FacilityID: facilityID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String("analytics-processing"),
		Payload:        payloadBytes,
		InvocationType: "Event",
	}

	_, err = c.svc.Invoke(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to invoke Lambda: %w", err)
	}

	return nil
}
