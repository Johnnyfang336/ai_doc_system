package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	
	_ "github.com/lib/pq"
	"ai-doc-system/internal/config"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	return db, nil
}

func Migrate(db *sql.DB) error {
	// Create migration record table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}
	
	// Read migration files
	migrationFiles, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return err
	}
	
	sort.Strings(migrationFiles)
	
	for _, file := range migrationFiles {
		version := strings.TrimSuffix(filepath.Base(file), ".sql")
		
		// Check if already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return err
		}
		
		if count > 0 {
			continue // Already applied, skip
		}
		
		// Read and execute migration file
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %v", version, err)
		}
		
		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			return err
		}
		
		fmt.Printf("Applied migration: %s\n", version)
	}
	
	return nil
}