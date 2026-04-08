package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Open opens a MySQL connection using dsn, pings the server to confirm
// connectivity, and configures the connection pool.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.Open: sql.Open: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Open: ping: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// RunMigrations applies all pending "up" migrations found at migrationsPath.
// A migrationsPath of "file://path/to/migrations" is expected.
// ErrNoChange is treated as success.
//
// dsn is used to open a separate connection with multiStatements=true, which
// is required by golang-migrate to execute multi-statement SQL files.
func RunMigrations(db *sql.DB, migrationsPath, dsn string) error {
	// golang-migrate requires multiStatements=true to run files with multiple statements.
	multiDSN := dsn
	if strings.Contains(dsn, "?") {
		multiDSN = dsn + "&multiStatements=true"
	} else {
		multiDSN = dsn + "?multiStatements=true"
	}
	migDB, err := sql.Open("mysql", multiDSN)
	if err != nil {
		return fmt.Errorf("db.RunMigrations: open migrate db: %w", err)
	}
	defer migDB.Close()

	driver, err := migratemysql.WithInstance(migDB, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("db.RunMigrations: create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("db.RunMigrations: new migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db.RunMigrations: up: %w", err)
	}

	return nil
}
