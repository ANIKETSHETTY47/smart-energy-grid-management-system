package repository

import (
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/domain"
	"github.com/jmoiron/sqlx"
)

// Repos - basically our db wrapper, no cap
type Repos struct {
	db *sqlx.DB
}

// New - creates a new repo instance, it's giving constructor vibes
func New(db *sqlx.DB) *Repos { return &Repos{db: db} }

// ListFacilities - gets all facilities, periodt
func (r *Repos) ListFacilities() ([]domain.Facility, error) {
	var out []domain.Facility
	err := r.db.Select(&out, `SELECT id, name FROM facilities ORDER BY id`)
	return out, err
}

// ListMeters - pulls all meters, slay
func (r *Repos) ListMeters() ([]domain.Meter, error) {
	var out []domain.Meter
	err := r.db.Select(&out, `SELECT id, facility_id, serial FROM meters ORDER BY id`)
	return out, err
}

// InsertReading - yeets a new reading into the db
func (r *Repos) InsertReading(rd *domain.Reading) error {
	_, err := r.db.Exec(`INSERT INTO readings(meter_id, timestamp, voltage, current, power_kw) VALUES ($1,$2,$3,$4,$5)`,
		rd.MeterID, rd.Timestamp, rd.Voltage, rd.Current, rd.PowerKW)
	return err
}
