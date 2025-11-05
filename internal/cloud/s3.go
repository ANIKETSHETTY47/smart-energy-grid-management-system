package cloud

import (
	"bytes"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client struct {
	svc    *s3.S3
	bucket string
}

func NewS3Client(region, bucket string) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return &S3Client{
		svc:    s3.New(sess),
		bucket: bucket,
	}, nil
}

// YOUR ORIGINAL CONTRIBUTION: Store reports in S3
func (c *S3Client) UploadReport(key string, data []byte) (string, error) {
	_, err := c.svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/pdf"),
	})

	if err != nil {
		return "", err
	}

	// Generate presigned URL
	req, _ := c.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(15 * time.Minute)
	return url, err
}
