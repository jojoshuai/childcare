package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLSleepStore is a MySQL-backed implementation of SleepStore.
type MySQLSleepStore struct {
	db *sql.DB
}

// NewSleepStore returns a SleepStore backed by the provided *sql.DB.
func NewSleepStore(db *sql.DB) SleepStore {
	return &MySQLSleepStore{db: db}
}

func (s *MySQLSleepStore) Create(r *model.SleepRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO sleep_records (id, child_id, start_time, end_time, woke_up, wake_count, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ChildID, r.StartTime, r.EndTime, r.WokeUp, r.WakeCount, r.CreatedBy, r.CreatedAt,
	)
	return err
}

func (s *MySQLSleepStore) GetByChildID(childID string) ([]*model.SleepRecord, error) {
	query := `SELECT id, child_id, start_time, end_time, woke_up, wake_count, created_by, created_at
	          FROM sleep_records WHERE child_id = ? ORDER BY start_time DESC`

	rows, err := s.db.Query(query, childID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.SleepRecord
	for rows.Next() {
		var r model.SleepRecord
		var endTime sql.NullTime
		if err := rows.Scan(
			&r.ID, &r.ChildID, &r.StartTime, &endTime,
			&r.WokeUp, &r.WakeCount, &r.CreatedBy, &r.CreatedAt,
		); err != nil {
			return nil, err
		}
		if endTime.Valid {
			r.EndTime = &endTime.Time
		}
		records = append(records, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *MySQLSleepStore) GetByID(id string) (*model.SleepRecord, error) {
	row := s.db.QueryRow(
		`SELECT id, child_id, start_time, end_time, woke_up, wake_count, created_by, created_at
		 FROM sleep_records WHERE id = ?`,
		id,
	)
	var r model.SleepRecord
	var endTime sql.NullTime
	err := row.Scan(
		&r.ID, &r.ChildID, &r.StartTime, &endTime,
		&r.WokeUp, &r.WakeCount, &r.CreatedBy, &r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if endTime.Valid {
		r.EndTime = &endTime.Time
	}
	return &r, nil
}

func (s *MySQLSleepStore) Update(r *model.SleepRecord) error {
	_, err := s.db.Exec(
		`UPDATE sleep_records SET start_time = ?, end_time = ?, woke_up = ?, wake_count = ? WHERE id = ?`,
		r.StartTime, r.EndTime, r.WokeUp, r.WakeCount, r.ID,
	)
	return err
}

func (s *MySQLSleepStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM sleep_records WHERE id = ?`, id)
	return err
}
