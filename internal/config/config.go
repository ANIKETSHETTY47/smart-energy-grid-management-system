package config

import "github.com/spf13/viper"

func Load() error {
	// API Configuration
	viper.SetDefault("API_ADDR", ":8080")

	// Database Configuration (keep for local dev)
	viper.SetDefault("DB_DSN", "postgres://postgres:postgres@localhost:5432/energy?sslmode=disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("MQTT_BROKER", "tcp://localhost:1883")

	// AWS Configuration
	viper.SetDefault("AWS_REGION", "us-east-1")
	viper.SetDefault("AWS_S3_BUCKET", "energy-grid-reports")
	viper.SetDefault("AWS_SNS_TOPIC_ARN", "")
	viper.SetDefault("USE_CLOUD_SERVICES", "false") // Toggle for local vs cloud

	viper.AutomaticEnv()
	return nil
}

func MQTTBroker() string     { return viper.GetString("MQTT_BROKER") }
func AWSRegion() string      { return viper.GetString("AWS_REGION") }
func S3Bucket() string       { return viper.GetString("AWS_S3_BUCKET") }
func SNSTopicArn() string    { return viper.GetString("AWS_SNS_TOPIC_ARN") }
func UseCloudServices() bool { return viper.GetBool("USE_CLOUD_SERVICES") }
