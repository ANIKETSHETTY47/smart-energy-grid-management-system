package main

import (
	"fmt"
	"time"

	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/aggregator"
	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/anomaly"
	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/converter"
	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/maintenance"
)

func main() {
	fmt.Println("=== Testing Energy Grid Analytics Library Integration ===\n")

	// Test 1: Energy Converter
	fmt.Println("1. Testing Energy Converter:")
	conv := &converter.EnergyConverter{}
	kwh := 5000.0
	mwh := conv.KWhToMWh(kwh)
	fmt.Printf("   %,.0f kWh = %.2f MWh\n", kwh, mwh)

	cost := conv.CalculateCost(100, 0.20, "peak")
	fmt.Printf("   Cost (peak): $%.2f\n", cost)

	efficiency := conv.CalculateEfficiency(120, 90)
	fmt.Printf("   Efficiency: %.2f%%\n\n", efficiency)

	// Test 2: Data Aggregator
	fmt.Println("2. Testing Data Aggregator:")
	points := []aggregator.Point{
		{Value: 10, Timestamp: time.Now().Add(-3 * time.Hour)},
		{Value: 15, Timestamp: time.Now().Add(-2 * time.Hour)},
		{Value: 20, Timestamp: time.Now().Add(-1 * time.Hour)},
		{Value: 25, Timestamp: time.Now()},
	}
	fmt.Printf("   Sum: %.2f\n", aggregator.Sum(points))
	fmt.Printf("   Average: %.2f\n", aggregator.Average(points))
	fmt.Printf("   Moving Average (window=2): %v\n\n", aggregator.MovingAverage(points, 2))

	// Test 3: Anomaly Detector
	fmt.Println("3. Testing Anomaly Detector:")
	readings := []anomaly.Reading{
		{Consumption: 10, Timestamp: 1},
		{Consumption: 11, Timestamp: 2},
		{Consumption: 10, Timestamp: 3},
		{Consumption: 200, Timestamp: 4}, // SPIKE!
		{Consumption: 12, Timestamp: 5},
		{Consumption: 13, Timestamp: 6},
	}
	detector := &anomaly.AnomalyDetector{Threshold: 2.0, WindowSize: 3}
	spikes := detector.DetectSpikes(readings)
	outliers := detector.DetectOutliers(readings)
	fmt.Printf("   Spikes detected: %d\n", len(spikes))
	fmt.Printf("   Outliers detected: %d\n\n", len(outliers))

	// Test 4: Maintenance Predictor
	fmt.Println("4. Testing Maintenance Predictor:")
	health := maintenance.AssetHealth{
		HoursRun:           2500,
		FailureRatePerYear: 0.3,
		LastService:        time.Now().Add(-180 * 24 * time.Hour),
		ServiceInterval:    365 * 24 * time.Hour,
	}
	risk := maintenance.FailureRisk(health.FailureRatePerYear, 90*24*time.Hour)
	nextService := maintenance.NextServiceDate(health)
	fmt.Printf("   Failure risk (90 days): %.2f%%\n", risk*100)
	fmt.Printf("   Next service date: %s\n", nextService.Format("2006-01-02"))

	fmt.Println("\n✅ All library integrations tested successfully!")
}
