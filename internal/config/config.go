package config

import "github.com/spf13/viper"

func Load() error {
	// Try to load .env file for local development
	viper.SetConfigFile(".env")
	viper.ReadInConfig() // Ignore error, will use env vars if .env doesn't exist

	// API Configuration
	viper.SetDefault("API_ADDR", ":5000")

	// Database Configuration (keep for local dev)
	viper.SetDefault("DB_DSN", "postgres://postgres:postgres@localhost:5432/energy?sslmode=disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("MQTT_BROKER", "tcp://localhost:1883")

	// AWS Configuration
	viper.SetDefault("AWS_REGION", "eu-north-1")
	viper.SetDefault("AWS_ACCESS_KEY_ID", "")
	viper.SetDefault("AWS_SECRET_ACCESS_KEY", "")
	
	// AWS Service Configuration
	viper.SetDefault("AWS_S3_BUCKET", "smart-energy-grid-reports-nci")
	viper.SetDefault("AWS_DYNAMODB_TABLE_READINGS", "energy-grid-readings")
	viper.SetDefault("AWS_DYNAMODB_TABLE_ALERTS", "energy-grid-alerts")
	viper.SetDefault("AWS_DYNAMODB_TABLE_EQUIPMENT", "energy-grid-equipment")
	viper.SetDefault("AWS_SNS_TOPIC_ARN", "")
	viper.SetDefault("USE_CLOUD_SERVICES", "true")

	viper.AutomaticEnv()
	return nil
}

// API Configuration
func APIAddr() string { return viper.GetString("API_ADDR") }

// Local Development Configuration
func MQTTBroker() string { return viper.GetString("MQTT_BROKER") }
func DBConnectionString() string { return viper.GetString("DB_DSN") }

// AWS Configuration
func AWSRegion() string           { return viper.GetString("AWS_REGION") }
func AWSAccessKeyID() string      { return viper.GetString("AWS_ACCESS_KEY_ID") }
func AWSSecretAccessKey() string  { return viper.GetString("AWS_SECRET_ACCESS_KEY") }

// AWS Service Configuration
func S3Bucket() string                  { return viper.GetString("AWS_S3_BUCKET") }
func DynamoDBTableReadings() string     { return viper.GetString("AWS_DYNAMODB_TABLE_READINGS") }
func DynamoDBTableAlerts() string       { return viper.GetString("AWS_DYNAMODB_TABLE_ALERTS") }
func DynamoDBTableEquipment() string    { return viper.GetString("AWS_DYNAMODB_TABLE_EQUIPMENT") }
func SNSTopicArn() string               { return viper.GetString("AWS_SNS_TOPIC_ARN") }
func UseCloudServices() bool            { return viper.GetBool("USE_CLOUD_SERVICES") }
