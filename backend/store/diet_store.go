package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLDietStore is a MySQL-backed implementation of DietStore.
type MySQLDietStore struct {
	db *sql.DB
}

// NewDietStore returns a DietStore backed by the provided *sql.DB.
func NewDietStore(db *sql.DB) DietStore {
	return &MySQLDietStore{db: db}
}

func (s *MySQLDietStore) Create(r *model.DietRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO diet_records (id, child_id, food_name, food_type, amount_level, record_time, notes, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ChildID, r.FoodName, r.FoodType, r.AmountLevel, r.RecordTime, r.Notes, r.CreatedBy, r.CreatedAt,
	)
	return err
}

func (s *MySQLDietStore) GetByChildID(childID string) ([]*model.DietRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, child_id, food_name, food_type, amount_level, record_time, notes, created_by, created_at
		 FROM diet_records WHERE child_id = ? ORDER BY record_time DESC`,
		childID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.DietRecord
	for rows.Next() {
		var r model.DietRecord
		var notes sql.NullString
		if err := rows.Scan(
			&r.ID, &r.ChildID, &r.FoodName, &r.FoodType, &r.AmountLevel,
			&r.RecordTime, &notes, &r.CreatedBy, &r.CreatedAt,
		); err != nil {
			return nil, err
		}
		if notes.Valid {
			r.Notes = &notes.String
		}
		records = append(records, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *MySQLDietStore) GetByID(id string) (*model.DietRecord, error) {
	row := s.db.QueryRow(
		`SELECT id, child_id, food_name, food_type, amount_level, record_time, notes, created_by, created_at
		 FROM diet_records WHERE id = ?`,
		id,
	)
	var r model.DietRecord
	var notes sql.NullString
	err := row.Scan(
		&r.ID, &r.ChildID, &r.FoodName, &r.FoodType, &r.AmountLevel,
		&r.RecordTime, &notes, &r.CreatedBy, &r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if notes.Valid {
		r.Notes = &notes.String
	}
	return &r, nil
}

func (s *MySQLDietStore) Update(r *model.DietRecord) error {
	_, err := s.db.Exec(
		`UPDATE diet_records SET food_name = ?, food_type = ?, amount_level = ?, record_time = ?, notes = ? WHERE id = ?`,
		r.FoodName, r.FoodType, r.AmountLevel, r.RecordTime, r.Notes, r.ID,
	)
	return err
}

func (s *MySQLDietStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM diet_records WHERE id = ?`, id)
	return err
}
