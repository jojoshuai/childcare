package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLSupplementStore is a MySQL-backed implementation of SupplementStore.
type MySQLSupplementStore struct {
	db *sql.DB
}

// NewSupplementStore returns a SupplementStore backed by the provided *sql.DB.
func NewSupplementStore(db *sql.DB) SupplementStore {
	return &MySQLSupplementStore{db: db}
}

func (s *MySQLSupplementStore) Create(r *model.SupplementRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO supplement_records (id, child_id, supplement_name, dose, taken_at, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ChildID, r.SupplementName, r.Dose, r.TakenAt, r.CreatedBy, r.CreatedAt,
	)
	return err
}

func (s *MySQLSupplementStore) GetByChildID(childID string) ([]*model.SupplementRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, child_id, supplement_name, dose, taken_at, created_by, created_at
		 FROM supplement_records WHERE child_id = ? ORDER BY taken_at DESC`,
		childID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.SupplementRecord
	for rows.Next() {
		var r model.SupplementRecord
		var dose sql.NullString
		if err := rows.Scan(
			&r.ID, &r.ChildID, &r.SupplementName, &dose,
			&r.TakenAt, &r.CreatedBy, &r.CreatedAt,
		); err != nil {
			return nil, err
		}
		if dose.Valid {
			r.Dose = &dose.String
		}
		records = append(records, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *MySQLSupplementStore) GetByID(id string) (*model.SupplementRecord, error) {
	row := s.db.QueryRow(
		`SELECT id, child_id, supplement_name, dose, taken_at, created_by, created_at
		 FROM supplement_records WHERE id = ?`,
		id,
	)
	var r model.SupplementRecord
	var dose sql.NullString
	err := row.Scan(
		&r.ID, &r.ChildID, &r.SupplementName, &dose,
		&r.TakenAt, &r.CreatedBy, &r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if dose.Valid {
		r.Dose = &dose.String
	}
	return &r, nil
}

func (s *MySQLSupplementStore) Update(r *model.SupplementRecord) error {
	_, err := s.db.Exec(
		`UPDATE supplement_records SET supplement_name = ?, dose = ?, taken_at = ? WHERE id = ?`,
		r.SupplementName, r.Dose, r.TakenAt, r.ID,
	)
	return err
}

func (s *MySQLSupplementStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM supplement_records WHERE id = ?`, id)
	return err
}
