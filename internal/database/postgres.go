package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"chat-app/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	*sql.DB
}

func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	var dsn string
	if cfg.Host != "supabase" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.Name,
			cfg.SSLMode,
		)
	} else {
		dsn = fmt.Sprintf(
			"%s",
			cfg.URL,
		)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.DB.PingContext(ctx)
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func RunMigrations(cfg config.DatabaseConfig) error {
	var dsn string
	if cfg.Host != "supabase" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.Name,
			cfg.SSLMode,
		)
	} else {
		dsn = fmt.Sprintf(
			"%s",
			cfg.URL,
		)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database for migrations: %w", err)
	}

	// Create migrations tracking table if it doesn't exist
	createMigrationTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	if _, err := db.Exec(createMigrationTableSQL); err != nil {
		return fmt.Errorf("failed to create migrations tracking table: %w", err)
	}

	migrations := []string{
		"001_initial_schema.sql",
	}

	for _, migration := range migrations {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migration).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration, err)
		}

		if count > 0 {
			fmt.Printf("Migration %s already applied, skipping\n", migration)
			continue
		}

		migrationSQL, err := migrationsFS.ReadFile(fmt.Sprintf("migrations/%s", migration))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migration, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migration, err)
		}

		if _, err := tx.Exec(string(migrationSQL)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration, err)
		}

		fmt.Printf("Migration %s applied successfully\n", migration)
	}

	return nil
}
