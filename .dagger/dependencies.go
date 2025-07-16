package main

import (
	"context"

	"github.com/cryptellation/sma/dagger/internal/dagger"
)

// DependenciesContainer provides containers for all test dependencies (e.g., Postgres, Redis, etc.).
// PostgresContainer returns a service running Postgres initialized for integration tests.
func PostgresContainer(ctx context.Context, dag *dagger.Client, sourceDir *dagger.Directory) *dagger.Service {
	initSQL := sourceDir.File("deployments/docker-compose/postgresql/cryptellation.sql")

	c := dag.Container().
		From("postgres:15-alpine").
		WithEnvVariable("POSTGRES_PASSWORD", "postgres").
		WithEnvVariable("POSTGRES_USER", "postgres").
		WithEnvVariable("PGUSER", "postgres").
		WithEnvVariable("PGPASSWORD", "postgres").
		WithEnvVariable("POSTGRES_DB", "postgres")

	c = c.WithMountedFile("/docker-entrypoint-initdb.d/cryptellation.sql", initSQL)
	c = c.WithExposedPort(5432)

	return c.AsService()
}
