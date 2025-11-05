package cloud

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSClient struct {
	svc      *sns.SNS
	topicArn string
}

func NewSNSClient(region, topicArn string) (*SNSClient, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return &SNSClient{
		svc:      sns.New(sess),
		topicArn: topicArn,
	}, nil
}

// YOUR ORIGINAL CONTRIBUTION: Send alert notifications
func (c *SNSClient) SendAlert(subject, message string) error {
	_, err := c.svc.Publish(&sns.PublishInput{
		TopicArn: aws.String(c.topicArn),
		Subject:  aws.String(subject),
		Message:  aws.String(message),
	})
	return err
}
