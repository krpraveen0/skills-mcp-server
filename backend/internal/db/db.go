package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

// DB wraps a sql.DB connection with helper methods.
type DB struct {
	*sql.DB
}

// New opens a SQLite database, creates the data directory if needed,
// and runs all pending migrations.
func New(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// modernc sqlite driver is registered as "sqlite"
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode%%3DWAL&_pragma=foreign_keys%%3Don&_pragma=busy_timeout%%3D5000", path)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite benefits from limited connections
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db := &DB{sqlDB}

	if err := db.runMigrations(path); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	log.Printf("[db] Connected to SQLite at %s", path)
	return db, nil
}

// runMigrations applies all pending SQL migrations.
func (d *DB) runMigrations(dbPath string) error {
	// Resolve absolute migrations path relative to the binary location
	migrationsPath := resolveMigrationsPath()

	driver, err := sqlite3.WithInstance(d.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migration init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}

	version, _, _ := m.Version()
	log.Printf("[db] Migration version: %d", version)
	return nil
}

// resolveMigrationsPath finds the migrations directory.
func resolveMigrationsPath() string {
	// Try environment variable first
	if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
		return p
	}
	// Default to /migrations (Docker) or ./migrations (local)
	if _, err := os.Stat("/migrations"); err == nil {
		return "/migrations"
	}
	return "migrations"
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.DB.Close()
}
