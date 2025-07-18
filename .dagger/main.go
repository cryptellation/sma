// A generated module for Sma functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/cryptellation/sma/dagger/internal/dagger"
)

const (
	dockerImageName = "ghcr.io/cryptellation/sma"
)

type Sma struct{}

// Publish a new release.
func (ci *Sma) PublishTag(
	ctx context.Context,
	sourceDir *dagger.Directory,
	user *string,
	token *dagger.Secret,
) error {
	// Create Git repo access
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
		User:   user,
		Token:  token,
	})
	if err != nil {
		return err
	}

	// Publish new tag
	return repo.PublishTagFromReleaseTitle(ctx)
}

// Check returns a container that runs the checker.
func (ci *Sma) Check(
	sourceDir *dagger.Directory,
) *dagger.Container {
	c := dag.Container().From("ghcr.io/cryptellation/checker")
	return ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir).
		WithExec([]string{"checker"})
}

// Generate returns a container that runs the code generator.
func (ci *Sma) Generate(
	sourceDir *dagger.Directory,
) *dagger.Container {
	c := dag.Container().From("golang:" + goVersion() + "-alpine")
	return ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir).
		WithExec([]string{"sh", "-c", "go generate ./... && sh scripts/check-generation.sh"})
}

// Lint runs golangci-lint on the main repo (./...) only.
func (ci *Sma) Lint(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().
		From("golangci/golangci-lint:v1.62.0").
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint"))

	c = ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir)

	// Lint main repo only
	c = c.WithExec([]string{"golangci-lint", "run", "--timeout", "10m", "./..."})

	return c
}

// LintDagger runs golangci-lint on the .dagger directory only.
func (ci *Sma) LintDagger(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().
		From("golangci/golangci-lint:v1.62.0").
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint"))

	c = ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir)

	// Lint .dagger directory using parent config and module context
	c = c.WithExec([]string{"sh", "-c", "cd .dagger && golangci-lint run --config ../.golangci.yml --timeout 10m ."})

	return c
}

// UnitTests returns a container that runs the unit tests.
func (ci *Sma) UnitTests(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().From("golang:" + goVersion() + "-alpine")
	return ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir).
		WithExec([]string{"sh", "-c",
			"go test -tags=unit ./... | grep -v 'no test files'",
		})
}

// dbIntegrationTests runs the integration tests for the database against a fresh Postgres container.
func (ci *Sma) dbIntegrationTests(sourceDir *dagger.Directory) *dagger.Container {
	pg := PostgresService(dag, sourceDir)
	dsn := "host=postgres user=cryptellation password=cryptellation dbname=sma sslmode=disable"
	c := dag.Container().
		From("golang:"+goVersion()+"-alpine").
		WithServiceBinding("postgres", pg).
		WithEnvVariable("SQL_DSN", dsn)
	c = ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir)
	return c.WithExec([]string{"go", "test", "-tags=integration", "./svc/db/..."})
}

// IntegrationTests returns all integration test containers for this service.
func (ci *Sma) IntegrationTests(sourceDir *dagger.Directory) []*dagger.Container {
	return []*dagger.Container{
		ci.dbIntegrationTests(sourceDir),
	}
}

// EndToEndTests runs the end-to-end tests with all required services (DB, Temporal, SMA, Candlesticks)
// and env variables.
func (ci *Sma) EndToEndTests(
	sourceDir *dagger.Directory,
	binanceApiKey *dagger.Secret, //nolint:revive,stylecheck
	binanceSecretKey *dagger.Secret,
) *dagger.Container {
	// Start shared Postgres service
	db := PostgresService(dag, sourceDir)

	// Start Temporal service (uses shared Postgres)
	temporal := TemporalService(dag, sourceDir, db)

	// Start Candlesticks service and bind it to the test container (uses shared Postgres)
	candlesticks := CandlesticksService(dag, sourceDir, db, temporal, binanceApiKey, binanceSecretKey)

	// Start SMA service and bind it to the test container (uses shared Postgres)
	sma := Runner(dag, sourceDir, temporal, db)

	c := dag.Container().From("golang:" + goVersion() + "-alpine")
	c = ci.withGoCodeAndCacheAsWorkDirectory(c, sourceDir).
		WithServiceBinding("temporal", temporal).
		WithServiceBinding("sma", sma).
		WithServiceBinding("candlesticks", candlesticks).
		WithEnvVariable("TEMPORAL_ADDRESS", "temporal:7233")

	return c.WithExec([]string{"go", "test", "-v", "-tags=e2e", "./test"})
}

// Container returns a container with the application built in it.
func (ci *Sma) Container(
	sourceDir *dagger.Directory,
	// +optional
	targetPlatform string,
) *dagger.Container {
	// Get running OS, if that's an OS unsupported by Docker, replace by Linu
	os := runtime.GOOS
	if os == "darwin" {
		os = "linux"
	}

	// Set default runner info and override by argument
	runnerInfo := GoRunnersInfo["linux/amd64"]
	if targetPlatform != "" {
		info, ok := GoRunnersInfo[targetPlatform]
		if ok {
			runnerInfo = info
		}
	}

	return sourceDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
		BuildArgs: []dagger.BuildArg{
			{Name: "BUILDPLATFORM", Value: os + "/" + runtime.GOARCH},
			{Name: "TARGETOS", Value: runnerInfo.OS},
			{Name: "TARGETARCH", Value: runnerInfo.Arch},
			{Name: "BUILDBASEIMAGE", Value: runnerInfo.BuildBaseImage},
			{Name: "TARGETBASEIMAGE", Value: runnerInfo.TargetBaseImage},
		},
		Platform:   dagger.Platform(runnerInfo.OS + "/" + runnerInfo.Arch),
		Dockerfile: "build/container/Dockerfile",
	})
}

func (ci *Sma) PublishContainer(
	ctx context.Context,
	sourceDir *dagger.Directory,
) error {
	// Create Git repo access
	repo, err := NewGit(ctx, NewGitOptions{
		SrcDir: sourceDir,
	})
	if err != nil {
		return err
	}

	// Get tags
	tags, err := getDockerTags(ctx, repo)
	if err != nil {
		return err
	}

	return ci.publishContainer(ctx, sourceDir, tags)
}

func getDockerTags(ctx context.Context, repo Git) ([]string, error) {
	tags := make([]string, 0)

	// Generate last short sha
	lastShortSha, err := repo.GetLastCommitShortSHA(ctx)
	if err != nil {
		return nil, err
	}
	tags = append(tags, lastShortSha)

	// Stop here if this not main branch
	if name, err := repo.GetActualBranch(ctx); err != nil {
		return nil, err
	} else if name != "main" {
		return tags, nil
	}

	// Check if there is a new sem ver, if there is none, just stop here
	semVer, err := repo.GetLastTag(ctx)
	if err != nil {
		return nil, err
	} else if semVer == "" {
		return tags, nil
	}

	tags = append(tags, semVer)
	tags = append(tags, "latest")

	return tags, nil
}

// Publishes the worker docker image.
func (ci *Sma) publishContainer(
	ctx context.Context,
	sourceDir *dagger.Directory,
	tags []string,
) error {
	// Get platforms availables
	availablePlatforms := AvailablePlatforms()

	// Get images for each platform
	platformVariants := make([]*dagger.Container, 0, len(availablePlatforms))
	for _, targetPlatform := range availablePlatforms {
		runner := ci.Container(sourceDir, targetPlatform)
		platformVariants = append(platformVariants, runner)
	}

	// Set publication options from images
	publishOpts := dagger.ContainerPublishOpts{
		PlatformVariants: platformVariants,
	}

	// Publish with tags
	for _, tag := range tags {
		addr := fmt.Sprintf("%s:%s", dockerImageName, tag)
		if _, err := dag.Container().Publish(ctx, addr, publishOpts); err != nil {
			return err
		}
	}

	return nil
}

func goVersion() string {
	return runtime.Version()[2:]
}

func (ci *Sma) withGoCodeAndCacheAsWorkDirectory(
	c *dagger.Container,
	sourceDir *dagger.Directory,
) *dagger.Container {
	containerPath := "/go/src/github.com/cryptellation/sma"
	return c.
		// Add Go caches
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("gobuild")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("gocache")).

		// Add source code
		WithMountedDirectory(containerPath, sourceDir).

		// Add workdir
		WithWorkdir(containerPath)
}
