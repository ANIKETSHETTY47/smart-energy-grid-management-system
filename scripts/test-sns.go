// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/config"
// 	"github.com/aws/aws-sdk-go-v2/service/sns"
// )

// func main() {
// 	ctx := context.Background()
// 	topicArn := "arn:aws:sns:us-east-1:402831945884:energy-grid-alerts" // Change this

// 	// Load AWS configuration
// 	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
// 	if err != nil {
// 		log.Fatalf("unable to load SDK config: %v", err)
// 	}

// 	client := sns.NewFromConfig(cfg)

// 	// Test 1: List topics
// 	fmt.Println("=== Test 1: List Topics ===")
// 	listResult, err := client.ListTopics(ctx, &sns.ListTopicsInput{})
// 	if err != nil {
// 		log.Fatalf("failed to list topics: %v", err)
// 	}
// 	fmt.Println("Topics:")
// 	for _, topic := range listResult.Topics {
// 		fmt.Printf("  - %s\n", *topic.TopicArn)
// 	}

// 	// Test 2: Send test notification
// 	fmt.Println("\n=== Test 2: Send Test Notification ===")
// 	publishResult, err := client.Publish(ctx, &sns.PublishInput{
// 		TopicArn: aws.String(topicArn),
// 		Subject:  aws.String("Test Alert from Energy Grid System"),
// 		Message:  aws.String("This is a test alert to verify SNS configuration.\n\nTimestamp: " + time.Now().Format(time.RFC3339)),
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to publish: %v", err)
// 	}
// 	fmt.Printf("✓ Message sent successfully. MessageId: %s\n", *publishResult.MessageId)
// 	fmt.Println("  Check your email for the notification!")

// 	fmt.Println("\n✓ All SNS tests passed!")
// }
