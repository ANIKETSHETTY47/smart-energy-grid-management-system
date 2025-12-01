package cloud

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client wraps AWS S3 for storing stuff in the cloud
type S3Client struct {
	svc    *s3.Client
	bucket string
	ctx    context.Context
}

// NewS3Client creates a fresh S3 client
func NewS3Client(region, bucket string) (*S3Client, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &S3Client{
		svc:    s3.NewFromConfig(cfg),
		bucket: bucket,
		ctx:    ctx,
	}, nil
}

// UploadReport yeets a PDF to S3 and gives you a shareable link
func (c *S3Client) UploadReport(key string, data []byte, contentType string) (string, error) {
	// Upload the report to S3
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"uploaded-at": time.Now().Format(time.RFC3339),
		},
	}

	_, err := c.svc.PutObject(c.ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Create a presigned URL for downloading
	presignClient := s3.NewPresignClient(c.svc)
	presignInput := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	presignResult, err := presignClient.PresignGetObject(c.ctx, presignInput, func(opts *s3.PresignOptions) {
		opts.Expires = 1 * time.Hour // Link expires in 1 hour fr
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

// UploadDataFile dumps raw data into S3 for later
func (c *S3Client) UploadDataFile(key string, data []byte) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	}

	_, err := c.svc.PutObject(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload data file: %w", err)
	}

	return nil
}

// DownloadFile grabs a file from S3
func (c *S3Client) DownloadFile(key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	result, err := c.svc.GetObject(c.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return buf.Bytes(), nil
}

// ListReports shows what reports you've got in the bucket
func (c *S3Client) ListReports(prefix string) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	}

	var keys []string
	paginator := s3.NewListObjectsV2Paginator(c.svc, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(c.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			keys = append(keys, aws.ToString(obj.Key))
		}
	}

	return keys, nil
}

// DeleteFile removes a file from S3 (no cap)
func (c *S3Client) DeleteFile(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	_, err := c.svc.DeleteObject(c.ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}
