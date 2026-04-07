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
		`INSERT INTO users (id, family_id, username, password_hash, wx_openid, nickname, role, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.FamilyID, u.Username, u.PasswordHash, u.WxOpenID, u.Nickname, u.Role, u.CreatedAt,
	)
	return err
}

func (s *MySQLUserStore) GetByUsername(username string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, family_id, username, password_hash, wx_openid, nickname, role, created_at
         FROM users WHERE username = ?`,
		username,
	)
	return scanUser(row)
}

func (s *MySQLUserStore) GetByOpenID(openid string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, family_id, username, password_hash, wx_openid, nickname, role, created_at
         FROM users WHERE wx_openid = ?`,
		openid,
	)
	return scanUser(row)
}

func (s *MySQLUserStore) GetByID(id string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, family_id, username, password_hash, wx_openid, nickname, role, created_at
         FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

func (s *MySQLUserStore) UpdateFamily(userID, familyID, role string) error {
	_, err := s.db.Exec(
		`UPDATE users SET family_id = ?, role = ? WHERE id = ?`,
		familyID, role, userID,
	)
	return err
}

// scanUser reads one row into a User struct, using sql.NullString for nullable fields.
// Returns nil, nil when no row is found.
func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var familyID, username, passwordHash, wxOpenID, role sql.NullString
	err := row.Scan(
		&u.ID,
		&familyID,
		&username,
		&passwordHash,
		&wxOpenID,
		&u.Nickname,
		&role,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if familyID.Valid {
		u.FamilyID = &familyID.String
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
	return &u, nil
}
