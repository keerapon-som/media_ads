package packages

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresql(host string, port int, user, password, dbName, sslMode string) *PostgresConfig {
	return &PostgresConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}
}

func (p *PostgresConfig) Connect() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host,
		p.Port,
		p.User,
		p.Password,
		p.DBName,
		p.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres connection: %w", err)
	}

	return db, nil
}

func Disconnect(db *sql.DB) error {
	if db == nil {
		return nil
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("close postgres connection: %w", err)
	}

	return nil
}
