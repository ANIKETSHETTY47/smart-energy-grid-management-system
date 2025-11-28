package domain

import "time"

type Facility struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type Meter struct {
	ID         int64  `db:"id" json:"id"`
	FacilityID int64  `db:"facility_id" json:"facility_id"`
	Serial     string `db:"serial" json:"serial"`
}

type Reading struct {
	ID        int64     `db:"id" json:"id"`
	MeterID   int64     `db:"meter_id" json:"meter_id"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	Voltage   float64   `db:"voltage" json:"voltage"`
	Current   float64   `db:"current" json:"current"`
	PowerKW   float64   `db:"power_kw" json:"power_kw"`
}
