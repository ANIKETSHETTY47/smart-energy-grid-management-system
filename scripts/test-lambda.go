package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	lambdaSdk "github.com/aws/aws-sdk-go-v2/service/lambda"
)

type AnalyticsPayload struct {
	Date       string `json:"date"`
	FacilityID string `json:"facility_id"`
}

func main() {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	client := lambdaSdk.NewFromConfig(cfg)

	// Test 1: List Lambda functions
	fmt.Println("=== Test 1: List Lambda Functions ===")
	listResult, err := client.ListFunctions(ctx, &lambdaSdk.ListFunctionsInput{})
	if err != nil {
		log.Fatalf("failed to list functions: %v", err)
	}
	fmt.Println("Functions:")
	for _, fn := range listResult.Functions {
		// FunctionName is *string; Runtime is an enum (not a pointer)
		fmt.Printf("  - %s (%s)\n", aws.ToString(fn.FunctionName), string(fn.Runtime))
	}

	// Test 2: Invoke analytics-processing Lambda
	fmt.Println("\n=== Test 2: Invoke Analytics Processing ===")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	payload := AnalyticsPayload{
		Date:       yesterday,
		FacilityID: "facility-001",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("failed to marshal payload: %v", err)
	}

	invokeResult, err := client.Invoke(ctx, &lambdaSdk.InvokeInput{
		FunctionName: aws.String("analytics-processing"),
		Payload:      payloadBytes,
		// Optional: InvocationType: types.InvocationTypeRequestResponse,
	})
	if err != nil {
		log.Fatalf("failed to invoke: %v", err)
	}

	if invokeResult.FunctionError != nil {
		log.Fatalf("function error: %s", aws.ToString(invokeResult.FunctionError))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(invokeResult.Payload, &response); err != nil {
		log.Fatalf("failed to unmarshal response: %v", err)
	}

	fmt.Println("✓ Lambda invoked successfully")
	fmt.Printf("Response: %+v\n", response)
	fmt.Println("\n✓ All Lambda tests passed!")
}
