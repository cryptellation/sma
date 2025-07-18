package main

import (
	"github.com/cryptellation/sma/dagger/internal/dagger"
)

// PostgresService returns a service running Postgres initialized for integration tests.
func PostgresService(dag *dagger.Client, sourceDir *dagger.Directory) *dagger.Service {
	// Get the directory containing all initialization SQL scripts
	initSQLDir := sourceDir.Directory("deployments/docker-compose/postgresql")

	// Create the Postgres container with its environment variables
	c := dag.Container().
		From("postgres:15-alpine").
		WithEnvVariable("POSTGRES_PASSWORD", "postgres").
		WithEnvVariable("POSTGRES_USER", "postgres").
		WithEnvVariable("PGUSER", "postgres").
		WithEnvVariable("PGPASSWORD", "postgres").
		WithEnvVariable("POSTGRES_DB", "postgres")

	// Mount the initialization SQL directory
	c = c.WithMountedDirectory("/docker-entrypoint-initdb.d", initSQLDir)

	// Expose the default Postgres port
	c = c.WithExposedPort(5432)

	return c.AsService()
}

// TemporalService returns a Temporal service configured for Postgres, mounting dynamic config,
// and waiting for Postgres.
func TemporalService(dag *dagger.Client, sourceDir *dagger.Directory, db *dagger.Service) *dagger.Service {
	// Build the Temporal container with the official temporal image
	container := dag.Container().From("temporalio/auto-setup:1.25")

	// Bind the shared Postgres service to the container
	container = container.WithServiceBinding("postgresql", db)
	container = container.WithEnvVariable("DB", "postgres12")
	container = container.WithEnvVariable("DB_PORT", "5432")
	container = container.WithEnvVariable("POSTGRES_USER", "temporal")
	container = container.WithEnvVariable("POSTGRES_PWD", "temporal")
	container = container.WithEnvVariable("POSTGRES_SEEDS", "postgresql")
	container = container.WithEnvVariable("BIND_ON_IP", "0.0.0.0")
	container = container.WithEnvVariable("TEMPORAL_BROADCAST_ADDRESS", "127.0.0.1")

	// Set the dynamic config file for Temporal
	configDir := sourceDir.Directory("deployments/docker-compose/temporal")
	container = container.WithEnvVariable("DYNAMIC_CONFIG_FILE_PATH", "config/dynamicconfig/development-sql.yaml")
	container = container.WithMountedDirectory("/etc/temporal/config/dynamicconfig", configDir)

	// Expose the Temporal frontend port
	container = container.WithExposedPort(7233)

	return container.AsService()
}

// CandlesticksService returns a candlesticks worker container as a Dagger service, using the provided
// Postgres and Temporal services.
func CandlesticksService(
	dag *dagger.Client,
	_ *dagger.Directory,
	db *dagger.Service,
	temporal *dagger.Service,
	binanceAPIKey *dagger.Secret,
	binanceSecretKey *dagger.Secret,
) *dagger.Service {
	// Build the candlesticks container with the official image
	container := dag.Container().From("ghcr.io/cryptellation/candlesticks")

	// Bind the shared Postgres service to the container
	container = container.WithServiceBinding("postgres", db)
	container = container.WithEnvVariable(
		"SQL_DSN",
		"host=postgres user=cryptellation password=cryptellation dbname=candlesticks sslmode=disable",
	)

	// Bind the shared Temporal service to the container
	container = container.WithServiceBinding("temporal", temporal)
	container = container.WithEnvVariable("TEMPORAL_ADDRESS", "temporal:7233")

	// Set the Binance API credentials
	container = container.WithSecretVariable("BINANCE_API_KEY", binanceAPIKey)
	container = container.WithSecretVariable("BINANCE_SECRET_KEY", binanceSecretKey)

	// Expose the candlesticks service port
	container = container.WithExposedPort(9000)

	return container.AsService(dagger.ContainerAsServiceOpts{
		Args: []string{"sh", "-c", `
			worker database migrate
			worker serve
		`},
	})
}
