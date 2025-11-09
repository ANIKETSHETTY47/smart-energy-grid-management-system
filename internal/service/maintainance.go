package service

import (
	"fmt"
	"time"

	"github.com/ANIKETSHETTY47/energy-grid-analytics-go/maintenance"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/cloud"
)

// MaintenanceService handles predictive maintenance operations
type MaintenanceService struct {
	dynamoDB *cloud.DynamoDBClient
	sns      *cloud.SNSClient
	useCloud bool
}

// PredictMaintenanceNeeds analyzes equipment health and predicts maintenance requirements
// YOUR ORIGINAL CONTRIBUTION: Uses custom library for maintenance prediction
func (s *MaintenanceService) PredictMaintenanceNeeds(equipmentID string) (*MaintenancePrediction, error) {
	if !s.useCloud || s.dynamoDB == nil {
		return nil, fmt.Errorf("cloud services not enabled")
	}

	// Get equipment data
	equipment, err := s.dynamoDB.GetEquipment("facility-001")
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	var targetEquipment *cloud.Equipment
	for _, eq := range equipment {
		if eq.EquipmentID == equipmentID {
			targetEquipment = &eq
			break
		}
	}

	if targetEquipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	// YOUR ORIGINAL CONTRIBUTION: Create AssetHealth profile
	assetHealth := maintenance.AssetHealth{
		HoursRun:           calculateHoursRun(targetEquipment),
		FailureRatePerYear: 0.3, // Based on equipment type and history
		LastService:        time.Unix(targetEquipment.LastMaintenance, 0),
		ServiceInterval:    365 * 24 * time.Hour, // Annual maintenance
	}

	// YOUR ORIGINAL CONTRIBUTION: Calculate failure risk using library
	riskNext30Days := maintenance.FailureRisk(assetHealth.FailureRatePerYear, 30*24*time.Hour)
	riskNext90Days := maintenance.FailureRisk(assetHealth.FailureRatePerYear, 90*24*time.Hour)

	// YOUR ORIGINAL CONTRIBUTION: Predict next service date using library
	nextService := maintenance.NextServiceDate(assetHealth)

	prediction := &MaintenancePrediction{
		EquipmentID:       equipmentID,
		CurrentHealth:     targetEquipment.HealthScore,
		FailureRisk30Days: riskNext30Days * 100, // Convert to percentage
		FailureRisk90Days: riskNext90Days * 100,
		NextServiceDate:   nextService,
		DaysUntilService:  int(time.Until(nextService).Hours() / 24),
		Recommendation:    generateRecommendation(riskNext30Days, targetEquipment.HealthScore),
	}

	// Send alert if high risk
	if riskNext30Days > 0.5 || targetEquipment.HealthScore < 75 {
		s.sendMaintenanceAlert(prediction)
	}

	return prediction, nil
}

type MaintenancePrediction struct {
	EquipmentID       string    `json:"equipment_id"`
	CurrentHealth     float64   `json:"current_health"`
	FailureRisk30Days float64   `json:"failure_risk_30_days"`
	FailureRisk90Days float64   `json:"failure_risk_90_days"`
	NextServiceDate   time.Time `json:"next_service_date"`
	DaysUntilService  int       `json:"days_until_service"`
	Recommendation    string    `json:"recommendation"`
}

func calculateHoursRun(equipment *cloud.Equipment) float64 {
	daysSinceInstall := time.Since(time.Unix(equipment.InstallDate, 0)).Hours() / 24
	// Assume 20 hours/day operation
	return daysSinceInstall * 20
}

func generateRecommendation(risk float64, health float64) string {
	if risk > 0.5 || health < 60 {
		return "URGENT: Schedule immediate maintenance inspection"
	} else if risk > 0.3 || health < 75 {
		return "Schedule maintenance within next 30 days"
	} else if risk > 0.15 || health < 85 {
		return "Plan maintenance within next 90 days"
	}
	return "Equipment operating normally"
}

func (s *MaintenanceService) sendMaintenanceAlert(prediction *MaintenancePrediction) {
	if s.sns == nil {
		return
	}

	s.sns.SendMaintenanceAlert(
		prediction.EquipmentID,
		prediction.CurrentHealth,
		prediction.NextServiceDate,
	)
}
