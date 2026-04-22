package repository

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupMySQLContainer(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	// Encontrar o caminho absoluto para as migrações
	// Como estamos em customer/internal/adapters/db/repository, as migrações estão 2 níveis acima em ../migrations
	migration1, err := filepath.Abs("../migrations/000001_initial_schema.up.sql")
	if err != nil {
		t.Fatalf("failed to get absolute path for migration 1: %v", err)
	}
	migration2, err := filepath.Abs("../migrations/000002_add_outbox.up.sql")
	if err != nil {
		t.Fatalf("failed to get absolute path for migration 2: %v", err)
	}

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("food_project"),
		mysql.WithUsername("root"),
		mysql.WithPassword("root"),
		mysql.WithScripts(migration1, migration2),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	connStr, err := mysqlContainer.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("failed to open database connection: %v", err)
	}

	// Wait for DB to be ready
	var pingErr error
	for i := 0; i < 10; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if pingErr != nil {
		t.Fatalf("database not ready: %v", pingErr)
	}

	teardown := func() {
		db.Close()
		mysqlContainer.Terminate(ctx)
	}

	return db, teardown
}

func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("CI") == "" && os.Getenv("DOCKER_HOST") == "" {
		// Apenas um check básico se o docker está disponível
		// Se falhar o RunContainer já vai dar erro, mas aqui evitamos latência desnecessária se soubermos que não tem docker
	}
}
