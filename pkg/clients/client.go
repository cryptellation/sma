package clients

import (
	"context"

	"github.com/cryptellation/sma/api"
	temporalclient "go.temporal.io/sdk/client"
)

// Client is a client for the cryptellation sma service.
type Client interface {
	// List calls the list workflow.
	List(ctx context.Context, params api.ListWorkflowParams) (api.ListWorkflowResults, error)
	// Info calls the service info.
	Info(ctx context.Context) (api.ServiceInfoResults, error)
}

type client struct {
	temporal temporalclient.Client
}

// New creates a new client to execute temporal workflows.
func New(cl temporalclient.Client) Client {
	return &client{temporal: cl}
}

// List calls the list workflow.
func (c client) List(
	ctx context.Context,
	params api.ListWorkflowParams,
) (res api.ListWorkflowResults, err error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.ListWorkflowName, params)
	if err != nil {
		return api.ListWorkflowResults{}, err
	}

	// Get result and return
	err = exec.Get(ctx, &res)
	return res, err
}

// Info calls the service info.
func (c client) Info(ctx context.Context) (res api.ServiceInfoResults, err error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.ServiceInfoWorkflowName)
	if err != nil {
		return api.ServiceInfoResults{}, err
	}

	// Get result and return
	err = exec.Get(ctx, &res)
	return res, err
}
