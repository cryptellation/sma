package api

import (
	"time"

	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/candlesticks/pkg/period"
)

const (
	// WorkerTaskQueueName is the name of the task queue for the cryptellation worker.
	WorkerTaskQueueName = "CryptellationSmaTaskQueue"
)

const (
	// ListWorkflowName is the name of the workflow to list SMA points.
	ListWorkflowName = "ListWorkflow"
)

type (
	// ListWorkflowParams is the parameters of the List workflow.
	ListWorkflowParams struct {
		Exchange     string
		Pair         string
		Period       period.Symbol
		Start        time.Time
		End          time.Time
		PeriodNumber int
		PriceType    candlestick.PriceType
	}

	// ListWorkflowResults is the result of the List workflow.
	ListWorkflowResults struct {
		Data []struct {
			Time  time.Time
			Value float64
		}
	}
)

const (
	// ServiceInfoWorkflowName is the name of the workflow to get the service info.
	ServiceInfoWorkflowName = "ServiceInfoWorkflow"
)

type (
	// ServiceInfoParams contains the parameters of the service info workflow.
	ServiceInfoParams struct{}

	// ServiceInfoResults contains the result of the service info workflow.
	ServiceInfoResults struct {
		Version string
	}
)
