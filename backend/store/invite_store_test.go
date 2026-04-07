package store

import (
	"database/sql"
	"testing"
	"time"

	"childcare-backend/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestInviteStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO invite_codes").
		WithArgs("iid-1", "fam-1", "ABC123", sqlmock.AnyArg(), false, "uid-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := NewInviteStore(db)
	err = s.Create(&model.InviteCode{
		ID:        "iid-1",
		FamilyID:  "fam-1",
		Code:      "ABC123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedBy: "uid-1",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInviteStore_GetByCode_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "family_id", "code", "expires_at", "used", "created_by", "created_at",
	}).AddRow("iid-1", "fam-1", "ABC123", now.Add(24*time.Hour), false, "uid-1", now)

	mock.ExpectQuery("SELECT .* FROM invite_codes WHERE code").
		WithArgs("ABC123").
		WillReturnRows(rows)

	s := NewInviteStore(db)
	ic, err := s.GetByCode("ABC123")
	if err != nil {
		t.Fatal(err)
	}
	if ic == nil {
		t.Fatal("expected invite code, got nil")
	}
	if ic.Code != "ABC123" {
		t.Fatalf("expected code ABC123, got %s", ic.Code)
	}
	if ic.Used {
		t.Fatal("expected Used to be false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInviteStore_GetByCode_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM invite_codes WHERE code").
		WithArgs("XXXXXX").
		WillReturnError(sql.ErrNoRows)

	s := NewInviteStore(db)
	ic, err := s.GetByCode("XXXXXX")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ic != nil {
		t.Fatal("expected nil invite code")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInviteStore_MarkUsed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE invite_codes SET used").
		WithArgs("iid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := NewInviteStore(db)
	err = s.MarkUsed("iid-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
