package postgresql

import (
	"database/sql"
	"fmt"

	"github.com/chat-socio/backend/configuration"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func Connect(postgresConfig *configuration.PostgresConfig) (*sql.DB, error) {
	// Build the connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		postgresConfig.Host,
		postgresConfig.Port,
		postgresConfig.Username,
		postgresConfig.Password,
		postgresConfig.Database,
		postgresConfig.SSLMode,
	)

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}
