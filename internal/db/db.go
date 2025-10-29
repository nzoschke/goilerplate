package db

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func Init(driver, connection string) (*sqlx.DB, error) {
	// SQLite: create data directory if needed
	if driver == "sqlite" {
		dir := filepath.Dir(connection)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	db, err := sqlx.Connect(driver, connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Connection pool configuration (good defaults for all drivers)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	slog.Info("database connected", "driver", driver)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func Close(db *sqlx.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
