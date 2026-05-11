package store

import (
	"database/sql"
	"errors"

	"childcare-backend/model"
)

// MySQLUserStore is a MySQL-backed implementation of UserStore.
type MySQLUserStore struct {
	db *sql.DB
}

// NewUserStore returns a UserStore backed by the provided *sql.DB.
func NewUserStore(db *sql.DB) UserStore {
	return &MySQLUserStore{db: db}
}

func (s *MySQLUserStore) Create(u *model.User) error {
	_, err := s.db.Exec(
		`INSERT INTO users (id, username, password_hash, wx_openid, nickname, created_at)
         VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, u.WxOpenID, u.Nickname, u.CreatedAt,
	)
	return err
}

func (s *MySQLUserStore) GetByUsername(username string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, wx_openid, nickname, created_at
         FROM users WHERE username = ?`,
		username,
	)
	return scanUser(row)
}

func (s *MySQLUserStore) GetByOpenID(openid string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, wx_openid, nickname, created_at
         FROM users WHERE wx_openid = ?`,
		openid,
	)
	return scanUser(row)
}

func (s *MySQLUserStore) GetByID(id string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, wx_openid, nickname, created_at
         FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

// scanUser reads one row into a User struct.
func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var username, passwordHash, wxOpenID sql.NullString
	err := row.Scan(
		&u.ID,
		&username,
		&passwordHash,
		&wxOpenID,
		&u.Nickname,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
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
	return &u, nil
}
