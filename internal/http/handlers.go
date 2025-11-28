package http

import (
	"time"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/cloud"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/service"
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, svcs *service.Services) {
	g := app.Group("/")

	// NEW: Root + health
	g.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "smart-energy-grid-api",
			"status":  "ok",
			"endpoints": []string{
				"/health",
				"/facilities",
				"/meters",
				"/readings/recent?facility_id=facility-001&hours=24",
				"/alerts?facility_id=facility-001",
				"/alerts/:alert_id/acknowledge",
				"/analytics/generate",
				"/readings/check-anomaly",
			},
		})
	})
	g.Get("health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "time": time.Now().UTC()})
	})
	// NEW: Predictive maintenance endpoint
	g.Get("equipment/:id/maintenance", func(c *fiber.Ctx) error {
		equipmentID := c.Params("id")

		prediction, err := svcs.Maintenance.PredictMaintenanceNeeds(equipmentID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(prediction)
	})
	// Existing handlers
	g.Get("facilities", func(c *fiber.Ctx) error {
		items, err := svcs.Repos.ListFacilities()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(items)
	})

	g.Get("meters", func(c *fiber.Ctx) error {
		items, err := svcs.Repos.ListMeters()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(items)
	})

	// Trigger daily analytics via Lambda
	g.Post("analytics/generate", func(c *fiber.Ctx) error {
		type Request struct {
			FacilityID string `json:"facility_id"`
			Date       string `json:"date"` // YYYY-MM-DD (UTC)
		}

		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if req.FacilityID == "" {
			req.FacilityID = "facility-001"
		}
		// CHANGED: default empty date to TODAY (UTC) instead of yesterday
		if req.Date == "" {
			req.Date = time.Now().UTC().Format("2006-01-02")
		}

		reportURL, err := svcs.Analytics.GenerateDailyReport(req.FacilityID, req.Date)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error(), "date": req.Date})
		}

		// If Lambda returned no URL, surface a helpful message
		if reportURL == "" {
			return c.Status(200).JSON(fiber.Map{
				"message":  "Analytics processed, but no report URL returned (likely no data for the date).",
				"date":     req.Date,
				"facility": req.FacilityID,
			})
		}

		return c.JSON(fiber.Map{
			"message":    "Analytics generated successfully",
			"report_url": reportURL,
			"date":       req.Date,
			"facility":   req.FacilityID,
		})
	})

	// Get recent readings from DynamoDB
	g.Get("readings/recent", func(c *fiber.Ctx) error {
		facilityID := c.Query("facility_id", "facility-001")
		hours := c.QueryInt("hours", 24)

		readings, err := svcs.Readings.GetRecentReadings(facilityID, time.Duration(hours)*time.Hour)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"facility_id": facilityID,
			"hours":       hours,
			"count":       len(readings),
			"readings":    readings,
		})
	})

	// Get alerts from DynamoDB
	g.Get("alerts", func(c *fiber.Ctx) error {
		facilityID := c.Query("facility_id", "facility-001")
		severity := c.Query("severity", "")

		var severityPtr *string
		if severity != "" {
			severityPtr = &severity
		}

		alerts, err := svcs.Alerts.GetAlerts(facilityID, severityPtr)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"facility_id": facilityID,
			"severity":    severity,
			"count":       len(alerts),
			"alerts":      alerts,
		})
	})

	// Acknowledge an alert
	g.Post("alerts/:alert_id/acknowledge", func(c *fiber.Ctx) error {
		alertID := c.Params("alert_id")

		if err := svcs.Alerts.AcknowledgeAlert(alertID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message":  "Alert acknowledged",
			"alert_id": alertID,
		})
	})

	// Trigger anomaly detection manually
	g.Post("readings/check-anomaly", func(c *fiber.Ctx) error {
		type Request struct {
			FacilityID string  `json:"facility_id"`
			MeterID    string  `json:"meter_id"`
			Voltage    float64 `json:"voltage"`
			Current    float64 `json:"current"`
			PowerKW    float64 `json:"power_kw"`
		}

		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if !svcs.UseCloud || svcs.Lambda == nil {
			return c.Status(503).JSON(fiber.Map{"error": "Cloud services not enabled"})
		}

		payload := cloud.AnomalyDetectionPayload{
			FacilityID: req.FacilityID,
			MeterID:    req.MeterID,
			Timestamp:  time.Now().Unix(),
			Voltage:    req.Voltage,
			Current:    req.Current,
			PowerKW:    req.PowerKW,
		}

		result, err := svcs.Lambda.InvokeAnomalyDetection(payload)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "Anomaly detection completed",
			"result":  result,
		})
	})

	// NEW: Friendly 404
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"error": "route not found",
			"hint":  "see GET / for available endpoints",
		})
	})
}
