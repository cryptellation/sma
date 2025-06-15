package svc

import (
	"time"

	candlesticksapi "github.com/cryptellation/candlesticks/api"
	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/sma/api"
	"github.com/cryptellation/sma/pkg/sma"
	"github.com/cryptellation/sma/svc/db"
	"github.com/cryptellation/timeseries"
	"go.temporal.io/sdk/workflow"
)

// ListSMAWorkflow returns the SMA points for a given pair and exchange.
func (wf *workflows) ListSMAWorkflow(
	ctx workflow.Context,
	params api.ListWorkflowParams,
) (api.ListWorkflowResults, error) {
	logger := workflow.GetLogger(ctx)

	// Process the params
	params.Start = params.Period.RoundTime(params.Start)
	params.End = params.Period.RoundTime(params.End)

	logger.Info("Got request for SMA",
		"start", params.Start,
		"end", params.End,
		"pair", params.Pair,
		"exchange", params.Exchange,
		"period", params.Period)

	// Get SMA from DB and check if it's up to date
	res, upToDate, err := wf.getSMAFromDBAndCheck(ctx, params)
	if err != nil {
		return api.ListWorkflowResults{}, err
	} else if upToDate {
		logger.Info("SMA is up to date, returning")
		return res, nil
	}

	// Generate and upsert SMA points
	logger.Info("SMA is outdated, invalid or missing points, recalculating")
	res, err = wf.generateAndUpsertSMA(ctx, params)
	if err != nil {
		return api.ListWorkflowResults{}, err
	}

	return res, err
}

func (wf *workflows) getSMAFromDBAndCheck(
	ctx workflow.Context,
	params api.ListWorkflowParams,
) (res api.ListWorkflowResults, upToDate bool, err error) {
	logger := workflow.GetLogger(ctx)

	// Get cached SMA from DB
	var readDBRes db.ReadSMAActivityResults
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		wf.db.ReadSMAActivity, db.ReadSMAActivityParams{
			Exchange:     params.Exchange,
			Pair:         params.Pair,
			Period:       params.Period,
			PeriodNumber: params.PeriodNumber,
			PriceType:    params.PriceType,
			Start:        params.Start,
			End:          params.End,
		}).Get(ctx, &readDBRes)
	if err != nil {
		return api.ListWorkflowResults{}, false, err
	}
	logger.Info("Got SMA points",
		"count", readDBRes.Data.Len())

	// Check if current candlestick will be requested
	// If that's the case, we'll need to recalculate the SMA as the value has changed
	requested := params.Period.RoundTime(params.End)
	roundedNow := params.Period.RoundTime(time.Now())
	possiblyOutdatedSMA := requested.Equal(roundedNow)

	// Check if the SMA is up to date
	missingPoints := readDBRes.Data.AreMissing(params.Start, params.End, params.Period.Duration(), 0)
	upToDate = !missingPoints && !possiblyOutdatedSMA && !sma.InvalidValues(readDBRes.Data)

	// Convert timeserie to slice of structs
	data := make([]struct {
		Time  time.Time
		Value float64
	}, 0, readDBRes.Data.Len())
	err = readDBRes.Data.Loop(func(t time.Time, v float64) (bool, error) {
		data = append(data, struct {
			Time  time.Time
			Value float64
		}{Time: t, Value: v})
		return false, nil
	})
	if err != nil {
		return api.ListWorkflowResults{}, false, err
	}

	return api.ListWorkflowResults{
		Data: data,
	}, upToDate, nil
}

func (wf *workflows) generateAndUpsertSMA(
	ctx workflow.Context,
	params api.ListWorkflowParams,
) (api.ListWorkflowResults, error) {
	// Get necessary candlesticks
	data, err := wf.generateSMA(ctx, params)
	if err != nil {
		return api.ListWorkflowResults{}, err
	}

	// Save SMA points to DB and return the result
	err = wf.upsertSMA(ctx, params, data)
	if err != nil {
		return api.ListWorkflowResults{}, err
	}

	return api.ListWorkflowResults{
		Data: data,
	}, nil
}

func (wf *workflows) generateSMA(
	ctx workflow.Context,
	params api.ListWorkflowParams,
) (*timeseries.TimeSerie[float64], error) {
	// Get necessary candlesticks
	start := params.Start.Add(-params.Period.Duration() * time.Duration(params.PeriodNumber))
	res, err := wf.candlesticks.ListCandlesticks(ctx, candlesticksapi.ListCandlesticksWorkflowParams{
		Exchange: params.Exchange,
		Pair:     params.Pair,
		Period:   params.Period,
		Start:    &start,
		End:      &params.End,
	}, &workflow.ChildWorkflowOptions{
		TaskQueue: candlesticksapi.WorkerTaskQueueName,
	})
	if err != nil {
		return nil, err
	}

	// Generate SMAs and return them
	csList := candlestick.NewList(params.Exchange, params.Pair, params.Period)
	for _, cs := range res.List {
		if err := csList.Set(cs); err != nil {
			return nil, err
		}
	}

	return sma.TimeSerie(sma.TimeSerieParams{
		Candlesticks: csList,
		PriceType:    params.PriceType,
		Start:        params.Start,
		End:          params.End,
		PeriodNumber: params.PeriodNumber,
	})
}

func (wf *workflows) upsertSMA(
	ctx workflow.Context,
	params api.ListWorkflowParams,
	data *timeseries.TimeSerie[float64],
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Upserting SMA points",
		"count", data.Len())

	var upsertDBRes db.UpsertSMAActivityResults
	return workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		wf.db.UpsertSMAActivity, db.UpsertSMAActivityParams{
			Exchange:     params.Exchange,
			Pair:         params.Pair,
			Period:       params.Period,
			PeriodNumber: params.PeriodNumber,
			PriceType:    params.PriceType,
			TimeSerie:    data,
		}).Get(ctx, &upsertDBRes)
}
