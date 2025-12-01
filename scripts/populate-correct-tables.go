package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// CORRECT TABLE NAMES matching your application
const (
	TableReadings  = "energy-grid-readings"
	TableAlerts    = "energy-grid-alerts"
	TableEquipment = "energy-grid-equipment"
	TableAnalytics = "AnalyticsSummaries" // Keep this one as is
)

func main() {
	ctx := context.Background()

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if accessKey == "" || secretKey == "" {
		log.Fatal(`
❌ AWS credentials not set!

Run:
export AWS_ACCESS_KEY_ID="AKIAVK3PK4IJAZ2HKIPA"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

Then: go run populate-correct-tables.go
`)
	}

	if region == "" {
		region = "us-east-1"
	}

	fmt.Printf("🔑 Region: %s\n", region)
	fmt.Println("🔄 Loading AWS configuration...")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey, secretKey, "",
		)),
	)
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	fmt.Println("✓ Testing connection...")
	_, err = client.ListTables(ctx, &dynamodb.ListTablesInput{Limit: aws.Int32(1)})
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	fmt.Println("✓ Connected!")
	fmt.Println()

	fmt.Println("📊 Populating tables with CORRECT names...")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	tables := []struct {
		name    string
		popFunc func(context.Context, *dynamodb.Client, string) error
	}{
		{TableReadings, populateReadings},
		{TableAlerts, populateAlerts},
		{TableEquipment, populateEquipment},
		{TableAnalytics, populateAnalytics},
	}

	for _, table := range tables {
		fmt.Printf("📝 Populating %s...\n", table.name)
		if err := table.popFunc(ctx, client, table.name); err != nil {
			fmt.Printf("⚠️  Error: %v\n", err)
		} else {
			fmt.Printf("✅ %s done!\n", table.name)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("🎉 Complete!")
	fmt.Println()
	fmt.Println("📊 Populated:")
	fmt.Println("  ✓ ~500 readings")
	fmt.Println("  ✓ 50 alerts")
	fmt.Println("  ✓ 5 equipment")
	fmt.Println("  ✓ 60 analytics")
	fmt.Println()
	fmt.Printf("🌐 View: https://%s.console.aws.amazon.com/dynamodbv2/home?region=%s#tables\n", region, region)
}

type Reading struct {
	DeviceID    string  `dynamodbav:"device_id"`
	Timestamp   int64   `dynamodbav:"timestamp"`
	Voltage     float64 `dynamodbav:"voltage"`
	Current     float64 `dynamodbav:"current"`
	PowerKW     float64 `dynamodbav:"power_kw"`
	Status      string  `dynamodbav:"status"`
	Temperature float64 `dynamodbav:"temperature"`
}

type Alert struct {
	AlertID      string                 `dynamodbav:"alert_id"`
	CreatedAt    int64                  `dynamodbav:"created_at"`
	Severity     string                 `dynamodbav:"severity"`
	Type         string                 `dynamodbav:"type"`
	Message      string                 `dynamodbav:"message"`
	Acknowledged bool                   `dynamodbav:"acknowledged"`
	EquipmentID  string                 `dynamodbav:"equipment_id"`
	Metadata     map[string]interface{} `dynamodbav:"metadata"`
}

type Equipment struct {
	EquipmentID     string  `dynamodbav:"equipment_id"`
	Type            string  `dynamodbav:"type"`
	InstallDate     int64   `dynamodbav:"install_date"`
	LastMaintenance int64   `dynamodbav:"last_maintenance"`
	HealthScore     float64 `dynamodbav:"health_score"`
	Status          string  `dynamodbav:"status"`
}

type AnalyticsSummary struct {
	FacilityID          string             `dynamodbav:"facilityId"`
	Date                string             `dynamodbav:"date"`
	ReadingCount        int                `dynamodbav:"readingCount"`
	TotalConsumption    float64            `dynamodbav:"totalConsumption"`
	TotalConsumptionMWh float64            `dynamodbav:"totalConsumptionMWh"`
	AveragePower        float64            `dynamodbav:"averagePower"`
	PeakPower           float64            `dynamodbav:"peakPower"`
	MinPower            float64            `dynamodbav:"minPower"`
	AvgVoltage          float64            `dynamodbav:"avgVoltage"`
	VoltageStdDev       float64            `dynamodbav:"voltageStdDev"`
	AvgCurrent          float64            `dynamodbav:"avgCurrent"`
	PowerFactor         float64            `dynamodbav:"powerFactor"`
	PeakHour            string             `dynamodbav:"peakHour"`
	EstimatedCost       float64            `dynamodbav:"estimatedCost"`
	HourlyData          map[string]float64 `dynamodbav:"hourlyData"`
	CreatedAt           int64              `dynamodbav:"createdAt"`
}

func populateReadings(ctx context.Context, client *dynamodb.Client, tableName string) error {
	devices := []string{"device-001", "device-002", "device-003"}
	statuses := []string{"operational", "operational", "operational", "warning"}

	now := time.Now()
	rand.Seed(now.UnixNano())

	var writeRequests []types.WriteRequest
	count := 0

	for day := 0; day < 7; day++ {
		date := now.AddDate(0, 0, -day)
		for hour := 0; hour < 24; hour++ {
			for _, device := range devices {
				timestamp := date.Add(time.Duration(hour) * time.Hour).Unix()

				basePower := 1.0 + rand.Float64()*2.0
				if hour >= 9 && hour <= 17 {
					basePower *= 1.5
				}
				if rand.Float64() < 0.05 {
					basePower *= 2.5
				}

				reading := Reading{
					DeviceID:    device,
					Timestamp:   timestamp,
					Voltage:     220.0 + rand.Float64()*10.0,
					Current:     5.0 + rand.Float64()*3.0,
					PowerKW:     basePower,
					Status:      statuses[rand.Intn(len(statuses))],
					Temperature: 40.0 + rand.Float64()*15.0,
				}

				item, err := attributevalue.MarshalMap(reading)
				if err != nil {
					return err
				}

				writeRequests = append(writeRequests, types.WriteRequest{
					PutRequest: &types.PutRequest{Item: item},
				})
				count++

				if len(writeRequests) >= 25 {
					if err := batchWrite(ctx, client, tableName, writeRequests); err != nil {
						return err
					}
					fmt.Printf("   ✓ %d readings...\n", count)
					writeRequests = writeRequests[:0]
				}
			}
		}
	}

	if len(writeRequests) > 0 {
		batchWrite(ctx, client, tableName, writeRequests)
	}

	fmt.Printf("   ✓ Total: %d\n", count)
	return nil
}

func populateAlerts(ctx context.Context, client *dynamodb.Client, tableName string) error {
	severities := []string{"low", "medium", "high", "critical"}
	alertTypes := []string{"anomaly", "threshold", "equipment_failure", "maintenance"}
	equipmentIDs := []string{"device-001", "device-002", "device-003"}

	now := time.Now()
	rand.Seed(now.UnixNano())

	for i := 0; i < 50; i++ {
		daysAgo := rand.Intn(30)
		timestamp := now.AddDate(0, 0, -daysAgo).Unix()

		alert := Alert{
			AlertID:      fmt.Sprintf("alert-%d-%d", timestamp, rand.Intn(10000)),
			CreatedAt:    timestamp,
			Severity:     severities[rand.Intn(len(severities))],
			Type:         alertTypes[rand.Intn(len(alertTypes))],
			Message:      fmt.Sprintf("Power consumption spike: %.2f kW", 2.5+rand.Float64()*5.0),
			Acknowledged: rand.Float64() > 0.3,
			EquipmentID:  equipmentIDs[rand.Intn(len(equipmentIDs))],
			Metadata: map[string]interface{}{
				"power":     2.5 + rand.Float64()*5.0,
				"deviation": 50.0 + rand.Float64()*100.0,
			},
		}

		item, err := attributevalue.MarshalMap(alert)
		if err != nil {
			return err
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})
		if err != nil {
			return err
		}

		if (i+1)%10 == 0 {
			fmt.Printf("   ✓ %d alerts...\n", i+1)
		}
	}
	return nil
}

func populateEquipment(ctx context.Context, client *dynamodb.Client, tableName string) error {
	equipment := []Equipment{
		{
			EquipmentID:     "device-001",
			Type:            "smart_meter",
			InstallDate:     time.Now().AddDate(-2, 0, 0).Unix(),
			LastMaintenance: time.Now().AddDate(0, -6, 0).Unix(),
			HealthScore:     92.5,
			Status:          "operational",
		},
		{
			EquipmentID:     "device-002",
			Type:            "smart_meter",
			InstallDate:     time.Now().AddDate(-1, -6, 0).Unix(),
			LastMaintenance: time.Now().AddDate(0, -3, 0).Unix(),
			HealthScore:     88.3,
			Status:          "operational",
		},
		{
			EquipmentID:     "device-003",
			Type:            "smart_meter",
			InstallDate:     time.Now().AddDate(-3, 0, 0).Unix(),
			LastMaintenance: time.Now().AddDate(0, -8, 0).Unix(),
			HealthScore:     75.8,
			Status:          "warning",
		},
		{
			EquipmentID:     "transformer-01",
			Type:            "transformer",
			InstallDate:     time.Now().AddDate(-5, 0, 0).Unix(),
			LastMaintenance: time.Now().AddDate(0, -12, 0).Unix(),
			HealthScore:     82.0,
			Status:          "operational",
		},
		{
			EquipmentID:     "generator-01",
			Type:            "backup_generator",
			InstallDate:     time.Now().AddDate(-4, 0, 0).Unix(),
			LastMaintenance: time.Now().AddDate(0, -2, 0).Unix(),
			HealthScore:     95.2,
			Status:          "operational",
		},
	}

	for i, eq := range equipment {
		item, err := attributevalue.MarshalMap(eq)
		if err != nil {
			return err
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})
		if err != nil {
			return err
		}
		fmt.Printf("   ✓ %d/%d (%s)\n", i+1, len(equipment), eq.EquipmentID)
	}
	return nil
}

func populateAnalytics(ctx context.Context, client *dynamodb.Client, tableName string) error {
	facilities := []string{"facility-001", "facility-002"}
	now := time.Now()
	rand.Seed(now.UnixNano())

	count := 0
	for day := 0; day < 30; day++ {
		date := now.AddDate(0, 0, -day).Format("2006-01-02")

		for _, facility := range facilities {
			hourlyData := make(map[string]float64)
			for hour := 0; hour < 24; hour++ {
				power := 1.0 + rand.Float64()*3.0
				if hour >= 9 && hour <= 17 {
					power *= 1.5
				}
				hourlyData[fmt.Sprintf("%02d", hour)] = power
			}

			summary := AnalyticsSummary{
				FacilityID:          facility,
				Date:                date,
				ReadingCount:        72,
				TotalConsumption:    120.5 + rand.Float64()*50.0,
				TotalConsumptionMWh: 0.12 + rand.Float64()*0.05,
				AveragePower:        1.67 + rand.Float64()*0.5,
				PeakPower:           4.5 + rand.Float64()*1.5,
				MinPower:            0.8 + rand.Float64()*0.3,
				AvgVoltage:          220.5 + rand.Float64()*5.0,
				VoltageStdDev:       2.3 + rand.Float64()*1.0,
				AvgCurrent:          6.2 + rand.Float64()*1.5,
				PowerFactor:         0.92 + rand.Float64()*0.05,
				PeakHour:            fmt.Sprintf("%02d", 9+rand.Intn(8)),
				EstimatedCost:       24.10 + rand.Float64()*10.0,
				HourlyData:          hourlyData,
				CreatedAt:           time.Now().Unix(),
			}

			item, err := attributevalue.MarshalMap(summary)
			if err != nil {
				return err
			}

			_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
				TableName: aws.String(tableName),
				Item:      item,
			})
			if err != nil {
				return err
			}

			count++
			if count%10 == 0 {
				fmt.Printf("   ✓ %d summaries...\n", count)
			}
		}
	}
	return nil
}

func batchWrite(ctx context.Context, client *dynamodb.Client, tableName string, writeRequests []types.WriteRequest) error {
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: writeRequests,
		},
	}
	_, err := client.BatchWriteItem(ctx, input)
	return err
}
