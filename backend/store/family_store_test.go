package store

import (
	"database/sql"
	"testing"
	"time"

	"childcare-backend/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestFamilyStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO families").
		WithArgs("fam-1", "Smith Family", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := NewFamilyStore(db)
	err = s.Create(&model.Family{ID: "fam-1", Name: "Smith Family", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFamilyStore_GetByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
		AddRow("fam-1", "Smith Family", now)

	mock.ExpectQuery("SELECT id, name, created_at FROM families WHERE id").
		WithArgs("fam-1").
		WillReturnRows(rows)

	s := NewFamilyStore(db)
	f, err := s.GetByID("fam-1")
	if err != nil {
		t.Fatal(err)
	}
	if f == nil {
		t.Fatal("expected family, got nil")
	}
	if f.Name != "Smith Family" {
		t.Fatalf("expected name Smith Family, got %s", f.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFamilyStore_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, created_at FROM families WHERE id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	s := NewFamilyStore(db)
	f, err := s.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if f != nil {
		t.Fatal("expected nil family")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFamilyStore_GetMembers(t *testing.T) {
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
	}).
		AddRow("uid-1", fid, "alice", nil, nil, "Alice", role, now).
		AddRow("uid-2", fid, nil, nil, "wx-open-2", "Bob", "member", now)

	mock.ExpectQuery("SELECT .* FROM users WHERE family_id").
		WithArgs("fam-1").
		WillReturnRows(rows)

	s := NewFamilyStore(db)
	members, err := s.GetMembers("fam-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].Nickname != "Alice" {
		t.Fatalf("expected first member Alice, got %s", members[0].Nickname)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
