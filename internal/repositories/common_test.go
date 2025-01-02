package repositories

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"log"
	"passkeeper/internal/database"
	"testing"
)

func createTestcontainer(ctx context.Context, t *testing.T) (*database.DBPool, func()) {
	dbName := "test_passkeeper"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		//postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		//postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatal(err)
	}
	uri, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := database.NewDB(uri, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err = pool.Migrate(); err != nil {
		t.Fatal(err)
	}

	return pool, func() {
		pool.Close()
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}
}
