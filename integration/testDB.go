package integration

import (
	"context"
	"dmt/internal"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName     = "dmt_test"
	dbUser     = "testuser"
	dbPassword = "testpass"
)

type TestContainer struct {
	Container  testcontainers.Container
	ConnString string
	ctx        context.Context
}

func NewTestDB(ctx context.Context) (*TestContainer, error) {
	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	err = runMigrations(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	tc := &TestContainer{
		Container:  container,
		ConnString: connStr,
		ctx:        ctx,
	}

	return tc, nil
}

func (tc *TestContainer) Terminate() error {
	return tc.Container.Terminate(context.Background())
}

func (tc *TestContainer) GetConnectionPool() (*pgxpool.Pool, error) {
	return internal.ConnectDb(tc.ctx, tc.ConnString)
}

func (tc *TestContainer) ClearDB(t *testing.T) {
	ctx := t.Context()

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
