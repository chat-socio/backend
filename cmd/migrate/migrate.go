package migrate

import (
	"fmt"
	"log"
	"os"

	"github.com/chat-socio/backend/configuration"
	"github.com/chat-socio/backend/infrastructure/postgresql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Migrate() {
	db, err := postgresql.Connect(configuration.ConfigInstance.Postgres)
	if err != nil {
		log.Println("Error connecting to database:", err)
		os.Exit(1)
	}

	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Println("Error creating migration driver:", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Println("Error creating new migration instance:", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Println("Error applying migrations:", err)
			os.Exit(1)
		}
	}

	fmt.Println("Migrations applied successfully")
}
