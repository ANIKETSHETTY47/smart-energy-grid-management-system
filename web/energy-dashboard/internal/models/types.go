package models

type Facility struct {
	FacilityID string `json:"facility_id"`
	Name       string `json:"name,omitempty"`
}

type Reading struct {
	Timestamp int64   `json:"timestamp"`
	PowerKW   float64 `json:"power_kw"`
	Voltage   float64 `json:"voltage"`
	Current   float64 `json:"current"`
}

type RecentReadingsResponse struct {
	Readings []Reading `json:"readings"`
}

type Alert struct {
	AlertID      string `json:"alertId"`
	FacilityID   string `json:"facilityId"`
	EquipmentID  string `json:"equipmentId"`
	Severity     string `json:"severity"`
	Type         string `json:"type"`
	Message      string `json:"message"`
	Timestamp    int64  `json:"timestamp"`
	Acknowledged bool   `json:"acknowledged"`
}

type AlertsResponse struct {
	Alerts []Alert `json:"alerts"`
}

type Health struct {
	Status string `json:"status"`
}

type AnalyticsGenerateRequest struct {
	FacilityID string `json:"facility_id"`
	Date       string `json:"date"`
}

type Analytics struct {
	TotalConsumption float64 `json:"total_consumption"`
	AveragePower     float64 `json:"average_power"`
	PeakPower        float64 `json:"peak_power"`
	PeakHour         string  `json:"peak_hour"`
}

type AnalyticsGenerateResponse struct {
	Date      string     `json:"date"`
	Facility  string     `json:"facility"`
	ReportURL string     `json:"report_url"`
	Analytics *Analytics `json:"analytics,omitempty"`
}

type Equipment struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	Status string  `json:"status"`
	Health float64 `json:"health"`
}
