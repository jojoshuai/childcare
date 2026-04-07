package store

import (
	"database/sql"
	"testing"
	"time"

	"childcare-backend/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestMeasurementStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO measurements").
		WithArgs(
			"mid-1", "cid-1", "weight", 12.5,
			sqlmock.AnyArg(), sqlmock.AnyArg(), "uid-1", sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := NewMeasurementStore(db)
	err = s.Create(&model.Measurement{
		ID:         "mid-1",
		ChildID:    "cid-1",
		Type:       "weight",
		Value:      12.5,
		MeasuredAt: time.Now(),
		CreatedBy:  "uid-1",
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurementStore_GetByChildID_NoFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "child_id", "type", "value", "measured_at", "note", "created_by", "created_at",
	}).
		AddRow("mid-1", "cid-1", "weight", 12.5, now, nil, "uid-1", now).
		AddRow("mid-2", "cid-1", "height", 85.0, now, "note text", "uid-1", now)

	mock.ExpectQuery("SELECT .* FROM measurements WHERE child_id").
		WithArgs("cid-1").
		WillReturnRows(rows)

	s := NewMeasurementStore(db)
	measurements, err := s.GetByChildID("cid-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 2 {
		t.Fatalf("expected 2 measurements, got %d", len(measurements))
	}
	if measurements[1].Note == nil || *measurements[1].Note != "note text" {
		t.Fatal("expected note to be set on second measurement")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurementStore_GetByChildID_WithFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "child_id", "type", "value", "measured_at", "note", "created_by", "created_at",
	}).AddRow("mid-1", "cid-1", "weight", 12.5, now, nil, "uid-1", now)

	mType := "weight"
	mock.ExpectQuery("SELECT .* FROM measurements WHERE child_id").
		WithArgs("cid-1", mType).
		WillReturnRows(rows)

	s := NewMeasurementStore(db)
	measurements, err := s.GetByChildID("cid-1", &mType)
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(measurements))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurementStore_GetByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "child_id", "type", "value", "measured_at", "note", "created_by", "created_at",
	}).AddRow("mid-1", "cid-1", "weight", 12.5, now, nil, "uid-1", now)

	mock.ExpectQuery("SELECT .* FROM measurements WHERE id").
		WithArgs("mid-1").
		WillReturnRows(rows)

	s := NewMeasurementStore(db)
	m, err := s.GetByID("mid-1")
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("expected measurement, got nil")
	}
	if m.Type != "weight" {
		t.Fatalf("expected type weight, got %s", m.Type)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurementStore_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM measurements WHERE id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	s := NewMeasurementStore(db)
	m, err := s.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if m != nil {
		t.Fatal("expected nil measurement")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurementStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE measurements SET type").
		WithArgs("height", 90.0, sqlmock.AnyArg(), sqlmock.AnyArg(), "mid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := NewMeasurementStore(db)
	err = s.Update(&model.Measurement{
		ID:         "mid-1",
		Type:       "height",
		Value:      90.0,
		MeasuredAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurementStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM measurements WHERE id").
		WithArgs("mid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := NewMeasurementStore(db)
	err = s.Delete("mid-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
