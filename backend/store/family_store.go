package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLFamilyStore is a MySQL-backed implementation of FamilyStore.
type MySQLFamilyStore struct {
	db *sql.DB
}

// NewFamilyStore returns a FamilyStore backed by the provided *sql.DB.
func NewFamilyStore(db *sql.DB) FamilyStore {
	return &MySQLFamilyStore{db: db}
}

func (s *MySQLFamilyStore) Create(f *model.Family) error {
	_, err := s.db.Exec(
		`INSERT INTO families (id, name, created_at) VALUES (?, ?, ?)`,
		f.ID, f.Name, f.CreatedAt,
	)
	return err
}

func (s *MySQLFamilyStore) GetByID(id string) (*model.Family, error) {
	row := s.db.QueryRow(
		`SELECT id, name, created_at FROM families WHERE id = ?`,
		id,
	)
	var f model.Family
	err := row.Scan(&f.ID, &f.Name, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (s *MySQLFamilyStore) GetMembers(familyID string) ([]*model.User, error) {
	rows, err := s.db.Query(
		`SELECT id, family_id, username, password_hash, wx_openid, nickname, role, created_at
         FROM users WHERE family_id = ?`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.User
	for rows.Next() {
		var u model.User
		var familyIDVal, username, passwordHash, wxOpenID, role sql.NullString
		if err := rows.Scan(
			&u.ID,
			&familyIDVal,
			&username,
			&passwordHash,
			&wxOpenID,
			&u.Nickname,
			&role,
			&u.CreatedAt,
		); err != nil {
			return nil, err
		}
		if familyIDVal.Valid {
			u.FamilyID = &familyIDVal.String
		}
		if username.Valid {
			u.Username = &username.String
		}
		if passwordHash.Valid {
			u.PasswordHash = &passwordHash.String
		}
		if wxOpenID.Valid {
			u.WxOpenID = &wxOpenID.String
		}
		if role.Valid {
			u.Role = &role.String
		}
		members = append(members, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}
