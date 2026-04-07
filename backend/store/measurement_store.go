package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLMeasurementStore is a MySQL-backed implementation of MeasurementStore.
type MySQLMeasurementStore struct {
	db *sql.DB
}

// NewMeasurementStore returns a MeasurementStore backed by the provided *sql.DB.
func NewMeasurementStore(db *sql.DB) MeasurementStore {
	return &MySQLMeasurementStore{db: db}
}

func (s *MySQLMeasurementStore) Create(m *model.Measurement) error {
	_, err := s.db.Exec(
		`INSERT INTO measurements (id, child_id, type, value, measured_at, note, created_by, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ChildID, m.Type, m.Value, m.MeasuredAt, m.Note, m.CreatedBy, m.CreatedAt,
	)
	return err
}

func (s *MySQLMeasurementStore) GetByChildID(childID string, measureType *string) ([]*model.Measurement, error) {
	query := `SELECT id, child_id, type, value, measured_at, note, created_by, created_at
              FROM measurements WHERE child_id = ?`
	args := []interface{}{childID}

	if measureType != nil {
		query += ` AND type = ?`
		args = append(args, *measureType)
	}
	query += ` ORDER BY measured_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measurements []*model.Measurement
	for rows.Next() {
		var m model.Measurement
		var note sql.NullString
		if err := rows.Scan(
			&m.ID, &m.ChildID, &m.Type, &m.Value,
			&m.MeasuredAt, &note, &m.CreatedBy, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		if note.Valid {
			m.Note = &note.String
		}
		measurements = append(measurements, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return measurements, nil
}

func (s *MySQLMeasurementStore) GetByID(id string) (*model.Measurement, error) {
	row := s.db.QueryRow(
		`SELECT id, child_id, type, value, measured_at, note, created_by, created_at
         FROM measurements WHERE id = ?`,
		id,
	)
	var m model.Measurement
	var note sql.NullString
	err := row.Scan(
		&m.ID, &m.ChildID, &m.Type, &m.Value,
		&m.MeasuredAt, &note, &m.CreatedBy, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if note.Valid {
		m.Note = &note.String
	}
	return &m, nil
}

func (s *MySQLMeasurementStore) Update(m *model.Measurement) error {
	_, err := s.db.Exec(
		`UPDATE measurements SET type = ?, value = ?, measured_at = ?, note = ? WHERE id = ?`,
		m.Type, m.Value, m.MeasuredAt, m.Note, m.ID,
	)
	return err
}

func (s *MySQLMeasurementStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM measurements WHERE id = ?`, id)
	return err
}
