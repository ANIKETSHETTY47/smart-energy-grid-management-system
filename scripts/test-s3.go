// package main

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/config"
// 	"github.com/aws/aws-sdk-go-v2/service/s3"
// )

// func main() {
// 	ctx := context.Background()
// 	bucketName := "energy-grid-reports" // Change this

// 	// Load AWS configuration
// 	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
// 	if err != nil {
// 		log.Fatalf("unable to load SDK config: %v", err)
// 	}

// 	client := s3.NewFromConfig(cfg)

// 	// Test 1: List buckets
// 	fmt.Println("=== Test 1: List Buckets ===")
// 	listResult, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
// 	if err != nil {
// 		log.Fatalf("failed to list buckets: %v", err)
// 	}
// 	fmt.Println("Buckets:")
// 	for _, bucket := range listResult.Buckets {
// 		fmt.Printf("  - %s\n", *bucket.Name)
// 	}

// 	// Test 2: Upload a test file
// 	fmt.Println("\n=== Test 2: Upload Test File ===")
// 	testContent := []byte("Test report generated at " + fmt.Sprint(time.Now()))
// 	key := "reports/test/test-report.txt"

// 	_, err = client.PutObject(ctx, &s3.PutObjectInput{
// 		Bucket:      aws.String(bucketName),
// 		Key:         aws.String(key),
// 		Body:        bytes.NewReader(testContent),
// 		ContentType: aws.String("text/plain"),
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to upload: %v", err)
// 	}
// 	fmt.Printf("✓ Successfully uploaded to s3://%s/%s\n", bucketName, key)

// 	// Test 3: Generate presigned URL
// 	fmt.Println("\n=== Test 3: Generate Presigned URL ===")
// 	presignClient := s3.NewPresignClient(client)
// 	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(key),
// 	}, func(opts *s3.PresignOptions) {
// 		opts.Expires = 1 * time.Hour
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to generate presigned URL: %v", err)
// 	}
// 	fmt.Println("✓ Presigned URL:", presignResult.URL)

// 	fmt.Println("\n✓ All S3 tests passed!")
// }
