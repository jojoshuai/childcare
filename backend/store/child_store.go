package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLChildStore is a MySQL-backed implementation of ChildStore.
type MySQLChildStore struct {
	db *sql.DB
}

// NewChildStore returns a ChildStore backed by the provided *sql.DB.
func NewChildStore(db *sql.DB) ChildStore {
	return &MySQLChildStore{db: db}
}

func (s *MySQLChildStore) Create(c *model.Child) error {
	_, err := s.db.Exec(
		`INSERT INTO children (id, family_id, name, gender, birth_date, created_at)
         VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID, c.FamilyID, c.Name, c.Gender, c.BirthDate, c.CreatedAt,
	)
	return err
}

func (s *MySQLChildStore) GetByFamilyID(familyID string) ([]*model.Child, error) {
	rows, err := s.db.Query(
		`SELECT id, family_id, name, gender, birth_date, created_at
         FROM children WHERE family_id = ?`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []*model.Child
	for rows.Next() {
		var c model.Child
		if err := rows.Scan(&c.ID, &c.FamilyID, &c.Name, &c.Gender, &c.BirthDate, &c.CreatedAt); err != nil {
			return nil, err
		}
		children = append(children, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return children, nil
}

func (s *MySQLChildStore) GetByID(id string) (*model.Child, error) {
	row := s.db.QueryRow(
		`SELECT id, family_id, name, gender, birth_date, created_at
         FROM children WHERE id = ?`,
		id,
	)
	var c model.Child
	err := row.Scan(&c.ID, &c.FamilyID, &c.Name, &c.Gender, &c.BirthDate, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *MySQLChildStore) Update(c *model.Child) error {
	_, err := s.db.Exec(
		`UPDATE children SET name = ?, gender = ?, birth_date = ? WHERE id = ?`,
		c.Name, c.Gender, c.BirthDate, c.ID,
	)
	return err
}

func (s *MySQLChildStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM children WHERE id = ?`, id)
	return err
}
