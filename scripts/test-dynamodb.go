// package main

// import (
// 	"context"
// 	"fmt"
// 	"time"
// 	"log"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/config"
// 	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
// 	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
// 	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
// )

// type TestReading struct {
// 	FacilityID string  `dynamodbav:"facilityId"`
// 	Timestamp  int64   `dynamodbav:"timestamp"`
// 	MeterID    string  `dynamodbav:"meterId"`
// 	Voltage    float64 `dynamodbav:"voltage"`
// 	Current    float64 `dynamodbav:"current"`
// 	PowerKW    float64 `dynamodbav:"powerKw"`
// }

// func main() {
// 	ctx := context.Background()

// 	// Load AWS configuration
// 	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
// 	if err != nil {
// 		log.Fatalf("unable to load SDK config: %v", err)
// 	}

// 	client := dynamodb.NewFromConfig(cfg)

// 	// Test 1: List tables
// 	fmt.Println("=== Test 1: List Tables ===")
// 	listResult, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
// 	if err != nil {
// 		log.Fatalf("failed to list tables: %v", err)
// 	}
// 	fmt.Println("Tables:", listResult.TableNames)

// 	// Test 2: Put a test reading
// 	fmt.Println("\n=== Test 2: Put Test Reading ===")
// 	testReading := TestReading{
// 		FacilityID: "facility-001",
// 		Timestamp:  time.Now().Unix(),
// 		MeterID:    "meter-test-001",
// 		Voltage:    220.5,
// 		Current:    5.2,
// 		PowerKW:    1.15,
// 	}

// 	item, err := attributevalue.MarshalMap(testReading)
// 	if err != nil {
// 		log.Fatalf("failed to marshal: %v", err)
// 	}

// 	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: aws.String("EnergyReadings"),
// 		Item:      item,
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to put item: %v", err)
// 	}
// 	fmt.Println("✓ Successfully stored test reading")

// 	// Test 3: Query the reading back
// 	fmt.Println("\n=== Test 3: Query Reading ===")
// 	queryResult, err := client.Query(ctx, &dynamodb.QueryInput{
// 		TableName:              aws.String("EnergyReadings"),
// 		KeyConditionExpression: aws.String("facilityId = :fid"),
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":fid": &types.AttributeValueMemberS{Value: "facility-001"},
// 		},
// 		Limit: aws.Int32(1),
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to query: %v", err)
// 	}

// 	var readings []TestReading
// 	err = attributevalue.UnmarshalListOfMaps(queryResult.Items, &readings)
// 	if err != nil {
// 		log.Fatalf("failed to unmarshal: %v", err)
// 	}

// 	fmt.Printf("✓ Retrieved %d readings\n", len(readings))
// 	if len(readings) > 0 {
// 		fmt.Printf("  Latest: %.2f kW at %v\n", readings[0].PowerKW, time.Unix(readings[0].Timestamp, 0))
// 	}

// 	fmt.Println("\n✓ All DynamoDB tests passed!")
// }