package bootstrap

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func parseDSNToMigrateURL(dsn string) string {
	parts := strings.Split(dsn, " ")
	params := make(map[string]string)

	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}

	user := params["user"]
	password := params["password"]
	host := params["host"]
	port := params["port"]
	dbname := params["dbname"]

	if port == "" {
		port = "5432"
	}

	url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	return url
}

func runMigrations(databaseURL string) error {
	migrateURL := parseDSNToMigrateURL(databaseURL)

	m, err := migrate.New("file://db/migrations", migrateURL)
	if err != nil {
		return fmt.Errorf("migration init failed: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations completed successfully")
	return nil
}
