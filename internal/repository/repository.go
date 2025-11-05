package repository

import (
	"github.com/jmoiron/sqlx"
	"smart/internal/domain"
)

type Repos struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repos { return &Repos{db: db} }

func (r *Repos) ListFacilities() ([]domain.Facility, error) {
	var out []domain.Facility
	err := r.db.Select(&out, `SELECT id, name FROM facilities ORDER BY id`)
	return out, err
}

func (r *Repos) ListMeters() ([]domain.Meter, error) {
	var out []domain.Meter
	err := r.db.Select(&out, `SELECT id, facility_id, serial FROM meters ORDER BY id`)
	return out, err
}

func (r *Repos) InsertReading(rd *domain.Reading) error {
	_, err := r.db.Exec(`INSERT INTO readings(meter_id, timestamp, voltage, current, power_kw) VALUES ($1,$2,$3,$4,$5)`,
		rd.MeterID, rd.Timestamp, rd.Voltage, rd.Current, rd.PowerKW)
	return err
}
