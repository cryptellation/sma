package svc

import (
	"github.com/cryptellation/sma/api"
	"github.com/cryptellation/version"
	"go.temporal.io/sdk/workflow"
)

const (
	// Version is the version of the service.
	Version = "devel"
	// CommitHash is the commit hash of the service.
	CommitHash = ""
)

func init() {
	version.SetVersion(Version)
	version.SetCommitHash(CommitHash)
}

// ServiceInfoWorkflow returns the service information.
func ServiceInfoWorkflow(_ workflow.Context, _ api.ServiceInfoParams) (api.ServiceInfoResults, error) {
	return api.ServiceInfoResults{
		Version: version.FullVersion(),
	}, nil
}
