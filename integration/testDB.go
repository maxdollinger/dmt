package integration

import (
	"context"
	"dmt/internal"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testAPIKey = "test-api-key-123"
	dbName     = "dmt_test"
	dbUser     = "testuser"
	dbPassword = "testpass"
)

var encodedAPIKey = base64.StdEncoding.EncodeToString([]byte(testAPIKey))

type TestContainer struct {
	Container  testcontainers.Container
	ConnString string
}

func SetupTestDB(t *testing.T) *TestContainer {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err, "Failed to start PostgreSQL container")

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get connection string")

	err = runMigrations(connStr)
	require.NoError(t, err, "Failed to run migrations")

	tc := &TestContainer{
		Container:  container,
		ConnString: connStr,
	}

	return tc
}

func (tc *TestContainer) Cleanup(t *testing.T) {
	if err := testcontainers.TerminateContainer(tc.Container); err != nil {
		t.Logf("Failed to terminate container: %v", err)
	}
}

func (tc *TestContainer) CreateApp(t *testing.T) (*fiber.App, *pgxpool.Pool) {
	ctx := context.Background()

	db, err := internal.ConnectDb(ctx, tc.ConnString)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	app := internal.CreateHttpServer(db, testAPIKey)

	return app, db
}

func (tc *TestContainer) ClearDB(t *testing.T) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, tc.ConnString)
	if err != nil {
		t.Fatalf("Failed to connect to database for cleanup: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "TRUNCATE TABLE device RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to clear database: %v", err)
	}
}

func SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+encodedAPIKey)
}

func runMigrations(connectionString string) error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	migrationsPath := "../internal/migrations"
	migrations, err := readMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	for _, migration := range migrations {
		if strings.HasSuffix(migration.Name, ".up.sql") {
			_, err := conn.Exec(ctx, migration.Content)
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.Name, err)
			}
		}
	}

	return nil
}

type Migration struct {
	Name    string
	Content string
}

func readMigrationFiles(dir string) ([]Migration, error) {
	var migrations []Migration

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		migrations = append(migrations, Migration{
			Name:    entry.Name(),
			Content: string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}
