package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	testContainer "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
)

var (
	testDB      *sql.DB
	dbContainer *testContainer.MySQLContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	dbContainer, err = testContainer.Run(ctx, "mysql:8.0",
		testContainer.WithDatabase("catalog_test"),
		testContainer.WithUsername("root"),
		testContainer.WithPassword("test"))

	if err != nil {
		log.Fatalf("failure on start db container: %v", err)
	}

	connString, err := dbContainer.ConnectionString(ctx, "parseTime=true&multiStatements=true")
	if err != nil {
		log.Fatalf("failure on open connection with db container: %v", err)
	}

	testDB, err = sql.Open("mysql", connString)
	if err != nil {
		log.Fatalf("failure on open db connection: %v", err)
	}

	runMigrations(testDB)

	code := m.Run()

	testDB.Close()
	dbContainer.Terminate(ctx)

	os.Exit(code)
}

func CountEventsInOutbox(t *testing.T, db *sql.DB, aggregateId string) int {
	var count int
	query := "SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ?"
	err := db.QueryRow(query, aggregateId).Scan(&count)
	if err != nil {
		t.Fatalf("error to count events in outbox table: %v", err)
	}
	return count
}

func GetLastEventPayload(t *testing.T, db *sql.DB, aggregateId string) map[string]interface{} {
	var payload []byte

	query := "SELECT payload FROM outbox_events WHERE aggregate_id = ? ORDER BY id DESC LIMIT 1"
	err := db.QueryRow(query, aggregateId).Scan(&payload)
	if err != nil {
		t.Fatalf("error to fetch payload in outbox: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("error to unmarshalling payload: %v", err)
	}
	return result
}

func insertRestaurant(t *testing.T, restaurant *restaurant.Restaurant) {
	restRepo := NewRestaurantRepository(testDB, NewOutboxRepository(testDB))
	err := restRepo.Save(context.Background(), restaurant)
	require.NoError(t, err)
}

func runMigrations(db *sql.DB) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("could not create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../../../internal/infra/db/migrate/migrations",
		"mysql",
		driver,
	)

	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run migrations: %v", err)
	}
}

func truncateTables(db *sql.DB) {
	tables := []string{"items", "categories", "menus", "outbox_events", "outbox_dlq", "restaurants"}

	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	for _, table := range tables {
		if _, err := db.Exec("TRUNCATE TABLE " + table); err != nil {
			log.Fatalf("failed to truncate %s: %v", table, err)
		}
	}
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}
