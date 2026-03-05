package util

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type Postgresql struct {
	db *sql.DB
}

// disable
func NewPostgresql(host string, port string, user string, password string, dbname string, sslMode string) (*Postgresql, error) {
	// Build the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslMode)

	// Open the database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return &Postgresql{
		db: db,
	}, nil
}

func (p *Postgresql) Close() error {
	if err := p.db.Close(); err != nil {
		return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
	}
	return nil
}

func (p *Postgresql) DB() *sql.DB {
	return p.db
}
