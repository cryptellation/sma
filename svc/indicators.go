package svc

import (
	"github.com/cryptellation/candlesticks/pkg/clients"
	"github.com/cryptellation/sma/api"
	"github.com/cryptellation/sma/svc/db"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// SMA is the interface for the SMA domain.
type SMA interface {
	Register(w worker.Worker)

	ListSMAWorkflow(
		ctx workflow.Context,
		params api.ListWorkflowParams,
	) (api.ListWorkflowResults, error)
}

// Check that the workflows implements the SMA interface.
var _ SMA = &workflows{}

type workflows struct {
	db           db.DB
	candlesticks clients.WfClient
}

// New creates a new SMA instance.
func New(db db.DB) SMA {
	return &workflows{
		candlesticks: clients.NewWfClient(),
		db:           db,
	}
}

// Register registers the workflows to the worker.
func (wf *workflows) Register(worker worker.Worker) {
	worker.RegisterWorkflowWithOptions(wf.ListSMAWorkflow, workflow.RegisterOptions{
		Name: api.ListWorkflowName,
	})

	worker.RegisterWorkflowWithOptions(ServiceInfoWorkflow, workflow.RegisterOptions{
		Name: api.ServiceInfoWorkflowName,
	})
}
