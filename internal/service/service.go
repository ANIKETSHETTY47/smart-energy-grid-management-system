package service

import (
	"encoding/json"
	"time"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/domain"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/repository"
	"github.com/jmoiron/sqlx"
)

type Services struct {
	Repos    *repository.Repos
	Readings *ReadingService
}

func New(db *sqlx.DB) *Services {
	repos := repository.New(db)
	return &Services{
		Repos:    repos,
		Readings: &ReadingService{repos: repos},
	}
}

type ReadingService struct {
	repos *repository.Repos
}

func (s *ReadingService) FromMQTT(topic string, payload []byte) error {
	var r struct {
		MeterID   string    `json:"meter_id"`
		Timestamp time.Time `json:"timestamp"`
		Voltage   float64   `json:"voltage"`
		Current   float64   `json:"current"`
		PowerKW   float64   `json:"power_kw"`
	}
	if err := json.Unmarshal(payload, &r); err != nil {
		return err
	}
	rd := &domain.Reading{
		MeterID:   1, // demo mapping
		Timestamp: r.Timestamp,
		Voltage:   r.Voltage,
		Current:   r.Current,
		PowerKW:   r.PowerKW,
	}
	return s.repos.InsertReading(rd)
}
