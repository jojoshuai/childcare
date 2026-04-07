package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLInviteStore is a MySQL-backed implementation of InviteStore.
type MySQLInviteStore struct {
	db *sql.DB
}

// NewInviteStore returns an InviteStore backed by the provided *sql.DB.
func NewInviteStore(db *sql.DB) InviteStore {
	return &MySQLInviteStore{db: db}
}

func (s *MySQLInviteStore) Create(ic *model.InviteCode) error {
	_, err := s.db.Exec(
		`INSERT INTO invite_codes (id, family_id, code, expires_at, used, created_by, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ic.ID, ic.FamilyID, ic.Code, ic.ExpiresAt, ic.Used, ic.CreatedBy, ic.CreatedAt,
	)
	return err
}

func (s *MySQLInviteStore) GetByCode(code string) (*model.InviteCode, error) {
	row := s.db.QueryRow(
		`SELECT id, family_id, code, expires_at, used, created_by, created_at
         FROM invite_codes WHERE code = ?`,
		code,
	)
	var ic model.InviteCode
	err := row.Scan(
		&ic.ID, &ic.FamilyID, &ic.Code,
		&ic.ExpiresAt, &ic.Used, &ic.CreatedBy, &ic.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ic, nil
}

func (s *MySQLInviteStore) MarkUsed(id string) error {
	_, err := s.db.Exec(`UPDATE invite_codes SET used = 1 WHERE id = ?`, id)
	return err
}
