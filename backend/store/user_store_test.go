package store

import (
	"database/sql"
	"testing"
	"time"

	"childcare-backend/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestUserStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO users").
		WithArgs(
			sqlmock.AnyArg(), // id
			sqlmock.AnyArg(), // family_id
			sqlmock.AnyArg(), // username
			sqlmock.AnyArg(), // password_hash
			sqlmock.AnyArg(), // wx_openid
			"Alice",          // nickname
			sqlmock.AnyArg(), // role
			sqlmock.AnyArg(), // created_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := NewUserStore(db)
	err = s.Create(&model.User{ID: "uuid-1", Nickname: "Alice", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserStore_GetByUsername_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "family_id", "username", "password_hash", "wx_openid", "nickname", "role", "created_at",
	}).AddRow("uid-1", nil, "alice", nil, nil, "Alice", nil, now)

	mock.ExpectQuery("SELECT .* FROM users WHERE username").
		WithArgs("alice").
		WillReturnRows(rows)

	s := NewUserStore(db)
	u, err := s.GetByUsername("alice")
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Nickname != "Alice" {
		t.Fatalf("expected nickname Alice, got %s", u.Nickname)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserStore_GetByUsername_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM users WHERE username").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	s := NewUserStore(db)
	u, err := s.GetByUsername("nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u != nil {
		t.Fatal("expected nil user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserStore_GetByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	fid := "fam-1"
	role := "owner"
	rows := sqlmock.NewRows([]string{
		"id", "family_id", "username", "password_hash", "wx_openid", "nickname", "role", "created_at",
	}).AddRow("uid-1", fid, nil, nil, nil, "Bob", role, now)

	mock.ExpectQuery("SELECT .* FROM users WHERE id").
		WithArgs("uid-1").
		WillReturnRows(rows)

	s := NewUserStore(db)
	u, err := s.GetByID("uid-1")
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.FamilyID == nil || *u.FamilyID != "fam-1" {
		t.Fatal("expected FamilyID to be set")
	}
	if u.Role == nil || *u.Role != "owner" {
		t.Fatal("expected Role to be owner")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserStore_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM users WHERE id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	s := NewUserStore(db)
	u, err := s.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u != nil {
		t.Fatal("expected nil user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserStore_GetByOpenID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM users WHERE wx_openid").
		WithArgs("ox999").
		WillReturnError(sql.ErrNoRows)

	s := NewUserStore(db)
	u, err := s.GetByOpenID("ox999")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u != nil {
		t.Fatal("expected nil user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserStore_UpdateFamily(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE users SET family_id").
		WithArgs("fam-1", "member", "uid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := NewUserStore(db)
	err = s.UpdateFamily("uid-1", "fam-1", "member")
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
