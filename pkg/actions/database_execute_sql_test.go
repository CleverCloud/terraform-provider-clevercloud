package actions

import (
	"context"
	"encoding/json"
	"testing"

	pgx "github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgresContainer(t *testing.T, ctx context.Context) (string, func()) {
	t.Helper()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	cleanup := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return connStr, cleanup
}

func TestPgSqlRowsToJson(t *testing.T) {
	ctx := t.Context()

	dsn, cleanup := setupPostgresContainer(t, ctx)
	defer cleanup()

	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err, "failed to connect to database")
	defer conn.Close(ctx)

	// Create a temporary test table
	_, err = conn.Exec(ctx, `
		CREATE TEMPORARY TABLE test_table (
			id INTEGER,
			name TEXT,
			price NUMERIC,
			active BOOLEAN
		)
	`)
	require.NoError(t, err, "failed to create test table")

	// Insert test data
	_, err = conn.Exec(ctx, `
		INSERT INTO test_table (id, name, price, active) VALUES
		(1, 'Product A', 19.99, true),
		(2, 'Product B', 29.50, false),
		(3, 'Product C', 9.99, true)
	`)
	require.NoError(t, err, "failed to insert test data")

	// Query the data
	rows, err := conn.Query(ctx, "SELECT id, name, price, active FROM test_table ORDER BY id")
	require.NoError(t, err, "failed to query test data")
	defer rows.Close()

	// Convert to JSON
	jsonData, err := PgSqlRowsToJson(rows)
	require.NoError(t, err, "PgSqlRowsToJson should not return an error")

	// Parse the JSON result
	var result []map[string]any
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err, "failed to unmarshal JSON result")

	// Verify the structure
	assert.Len(t, result, 3, "should have 3 rows")

	// Verify first row (JSON unmarshals numbers as float64)
	assert.Equal(t, float64(1), result[0]["id"], "first row id should be 1")
	assert.Equal(t, "Product A", result[0]["name"], "first row name should be Product A")
	assert.NotNil(t, result[0]["price"], "first row price should not be nil")
	assert.Equal(t, true, result[0]["active"], "first row active should be true")

	// Verify second row
	assert.Equal(t, float64(2), result[1]["id"], "second row id should be 2")
	assert.Equal(t, "Product B", result[1]["name"], "second row name should be Product B")
	assert.NotNil(t, result[1]["price"], "second row price should not be nil")
	assert.Equal(t, false, result[1]["active"], "second row active should be false")

	// Verify third row
	assert.Equal(t, float64(3), result[2]["id"], "third row id should be 3")
	assert.Equal(t, "Product C", result[2]["name"], "third row name should be Product C")
	assert.NotNil(t, result[2]["price"], "third row price should not be nil")
	assert.Equal(t, true, result[2]["active"], "third row active should be true")

	// Verify all rows have the expected columns
	for i, row := range result {
		assert.Contains(t, row, "id", "row %d should have id column", i)
		assert.Contains(t, row, "name", "row %d should have name column", i)
		assert.Contains(t, row, "price", "row %d should have price column", i)
		assert.Contains(t, row, "active", "row %d should have active column", i)
	}
}

func TestPgSqlRowsToJson_EmptyResult(t *testing.T) {
	ctx := context.Background()

	dsn, cleanup := setupPostgresContainer(t, ctx)
	defer cleanup()

	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err, "failed to connect to database")
	defer conn.Close(ctx)

	// Query with no results
	rows, err := conn.Query(ctx, "SELECT 1 AS id, 'test' AS name WHERE false")
	require.NoError(t, err, "failed to query")
	defer rows.Close()

	// Convert to JSON
	jsonData, err := PgSqlRowsToJson(rows)
	require.NoError(t, err, "PgSqlRowsToJson should not return an error")

	// Parse the JSON result
	var result []map[string]any
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err, "failed to unmarshal JSON result")

	// Verify empty array
	assert.Len(t, result, 0, "should have 0 rows")
}

func TestPgSqlRowsToJson_NullValues(t *testing.T) {
	ctx := context.Background()

	dsn, cleanup := setupPostgresContainer(t, ctx)
	defer cleanup()

	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err, "failed to connect to database")
	defer conn.Close(ctx)

	// Query with NULL values
	rows, err := conn.Query(ctx, "SELECT 1 AS id, NULL AS name, 'test' AS description")
	require.NoError(t, err, "failed to query")
	defer rows.Close()

	// Convert to JSON
	jsonData, err := PgSqlRowsToJson(rows)
	require.NoError(t, err, "PgSqlRowsToJson should not return an error")

	// Parse the JSON result
	var result []map[string]any
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err, "failed to unmarshal JSON result")

	// Verify the structure
	assert.Len(t, result, 1, "should have 1 row")
	assert.Equal(t, float64(1), result[0]["id"], "id should be 1")
	assert.Nil(t, result[0]["name"], "name should be nil")
	assert.Equal(t, "test", result[0]["description"], "description should be test")
}
