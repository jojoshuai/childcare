package store

import (
	"database/sql"
	"testing"
	"time"

	"childcare-backend/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestChildStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO children").
		WithArgs("cid-1", "fam-1", "Timmy", "male", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := NewChildStore(db)
	err = s.Create(&model.Child{
		ID:        "cid-1",
		FamilyID:  "fam-1",
		Name:      "Timmy",
		Gender:    "male",
		BirthDate: time.Now(),
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestChildStore_GetByFamilyID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "gender", "birth_date", "created_at"}).
		AddRow("cid-1", "fam-1", "Timmy", "male", now, now).
		AddRow("cid-2", "fam-1", "Sally", "female", now, now)

	mock.ExpectQuery("SELECT .* FROM children WHERE family_id").
		WithArgs("fam-1").
		WillReturnRows(rows)

	s := NewChildStore(db)
	children, err := s.GetByFamilyID("fam-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestChildStore_GetByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "gender", "birth_date", "created_at"}).
		AddRow("cid-1", "fam-1", "Timmy", "male", now, now)

	mock.ExpectQuery("SELECT .* FROM children WHERE id").
		WithArgs("cid-1").
		WillReturnRows(rows)

	s := NewChildStore(db)
	c, err := s.GetByID("cid-1")
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("expected child, got nil")
	}
	if c.Name != "Timmy" {
		t.Fatalf("expected name Timmy, got %s", c.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestChildStore_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM children WHERE id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	s := NewChildStore(db)
	c, err := s.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if c != nil {
		t.Fatal("expected nil child")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestChildStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE children SET name").
		WithArgs("Timmy Jr.", "male", sqlmock.AnyArg(), "cid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := NewChildStore(db)
	err = s.Update(&model.Child{ID: "cid-1", Name: "Timmy Jr.", Gender: "male", BirthDate: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestChildStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM children WHERE id").
		WithArgs("cid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := NewChildStore(db)
	err = s.Delete("cid-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
