package clients

// WfClient is a client for the cryptellation sma service from a workflow perspective.
type WfClient interface {
}

type wfClient struct{}

// NewWfClient creates a new workflow client.
// This client is used to call workflows from within other workflows.
// It is not used to call workflows from outside the workflow environment.
func NewWfClient() WfClient {
	return wfClient{}
}
