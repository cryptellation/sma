package main

import (
	"maps"
	"runtime"
	"slices"

	"github.com/cryptellation/sma/dagger/internal/dagger"
)

// RunnerInfo represents a Docker runner.
type RunnerInfo struct {
	OS              string
	Arch            string
	BuildBaseImage  string
	TargetBaseImage string
}

var (
	// GoRunnersInfo represents the different OS/Arch platform wanted for docker hub in Go service.
	GoRunnersInfo = map[string]RunnerInfo{
		"linux/386":      {OS: "linux", Arch: "386", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/amd64":    {OS: "linux", Arch: "amd64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm/v6":   {OS: "linux", Arch: "arm/v6", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm/v7":   {OS: "linux", Arch: "arm/v7", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/arm64/v8": {OS: "linux", Arch: "arm64/v8", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/ppc64le":  {OS: "linux", Arch: "ppc64le", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/riscv64":  {OS: "linux", Arch: "riscv64", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
		"linux/s390x":    {OS: "linux", Arch: "s390x", BuildBaseImage: "golang:alpine", TargetBaseImage: "alpine"},
	}
)

func AvailablePlatforms() []string {
	return slices.Collect(maps.Keys(GoRunnersInfo))
}

// Runner returns a container running the sma service built from the official Dockerfile,
// with its own Postgres and a given Temporal service.
func Runner(
	_ *dagger.Client,
	sourceDir *dagger.Directory,
	temporal *dagger.Service,
	db *dagger.Service,
) *dagger.Service {
	// Get the OS and architecture of the current machine
	os := runtime.GOOS
	if os == "darwin" {
		os = "linux"
	}
	arch := runtime.GOARCH

	// Get the runner info for the current platform
	runnerInfo := GoRunnersInfo["linux/amd64"]
	key := os + "/" + arch
	if info, ok := GoRunnersInfo[key]; ok {
		runnerInfo = info
	}

	// Build the container using the Dockerfile in the source directory
	container := sourceDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
		BuildArgs: []dagger.BuildArg{
			{Name: "BUILDPLATFORM", Value: os + "/" + arch},
			{Name: "TARGETOS", Value: runnerInfo.OS},
			{Name: "TARGETARCH", Value: runnerInfo.Arch},
			{Name: "BUILDBASEIMAGE", Value: runnerInfo.BuildBaseImage},
			{Name: "TARGETBASEIMAGE", Value: runnerInfo.TargetBaseImage},
		},
		Platform:   dagger.Platform(runnerInfo.OS + "/" + runnerInfo.Arch),
		Dockerfile: "build/container/Dockerfile",
	})

	// Bind the Postgres service to the container
	container = container.WithServiceBinding("postgres", db)
	container = container.WithEnvVariable(
		"SQL_DSN",
		"host=postgres user=cryptellation password=cryptellation dbname=sma sslmode=disable",
	)

	// Bind the Temporal service to the container
	container = container.WithServiceBinding("temporal", temporal)
	container = container.WithEnvVariable("TEMPORAL_ADDRESS", "temporal:7233")

	// Expose the default port (9000) as in Dockerfile
	container = container.WithExposedPort(9000)

	return container.AsService(dagger.ContainerAsServiceOpts{
		Args: []string{"sh", "-c", `
			worker database migrate
			worker serve
		`},
	})
}
