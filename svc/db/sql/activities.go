package sql

import (
	"context"
	"fmt"

	"github.com/cryptellation/sma/svc/db"
	"github.com/cryptellation/sma/svc/db/sql/entities"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostGres driver
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

var _ db.DB = (*Activities)(nil)

// Activities is a struct that contains all the methods to interact with the
// activities table in the database.
type Activities struct {
	db *sqlx.DB
}

// New creates a new activities.
func New(ctx context.Context, dsn string) (*Activities, error) {
	// Create embedded database access
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Create a structure
	a := &Activities{
		db: db,
	}

	return a, nil
}

// Register registers the activities.
func (a *Activities) Register(w worker.Worker) {
	w.RegisterActivityWithOptions(
		a.ReadSMAActivity,
		activity.RegisterOptions{Name: db.ReadSMAActivityName},
	)
	w.RegisterActivityWithOptions(
		a.UpsertSMAActivity,
		activity.RegisterOptions{Name: db.UpsertSMAActivityName},
	)
}

// Reset will reset the database.
func (a *Activities) Reset(ctx context.Context) error {
	_, err := a.db.ExecContext(ctx, "DELETE FROM sma")
	if err != nil {
		return fmt.Errorf("deleting sma rows: %w", err)
	}

	return nil
}

// ReadSMAActivity reads the SMA points from the database.
func (a *Activities) ReadSMAActivity(
	ctx context.Context,
	params db.ReadSMAActivityParams,
) (db.ReadSMAActivityResults, error) {
	// Query the SMA points
	rows, err := a.db.QueryxContext(
		ctx,
		`SELECT *
		FROM sma
		WHERE exchange = $1 AND 
			pair = $2 AND 
			period = $3 AND 
			period_number = $4 AND
			price_type = $5 AND
			time >= $6 AND time <= $7
		ORDER BY time ASC`,
		params.Exchange,
		params.Pair,
		params.Period,
		params.PeriodNumber,
		params.PriceType,
		params.Start.UTC(),
		params.End.UTC(),
	)
	if err != nil {
		return db.ReadSMAActivityResults{}, fmt.Errorf("querying SMA points: %w", err)
	}
	defer rows.Close()

	// Loop through the rows
	results := make([]entities.SimpleMovingAverage, 0)
	for rows.Next() {
		// Create the SMA point
		var point entities.SimpleMovingAverage
		err = rows.StructScan(&point)
		if err != nil {
			return db.ReadSMAActivityResults{}, fmt.Errorf("scanning SMA point: %w", err)
		}

		// Append the point
		results = append(results, point)
	}

	// To model list
	data, err := entities.FromEntityListToModelList(results)
	if err != nil {
		return db.ReadSMAActivityResults{}, fmt.Errorf("from entity list to model list: %w", err)
	}

	// Return the results
	return db.ReadSMAActivityResults{
		Data: data,
	}, nil
}

// UpsertSMAActivity upserts the SMA points in the database.
func (a *Activities) UpsertSMAActivity(
	ctx context.Context,
	params db.UpsertSMAActivityParams,
) (db.UpsertSMAActivityResults, error) {
	// Create entities
	ents, err := entities.FromModelListToEntityList(
		params.Exchange,
		params.Pair,
		params.Period,
		params.PeriodNumber,
		params.PriceType,
		params.TimeSerie)
	if err != nil {
		return db.UpsertSMAActivityResults{}, fmt.Errorf("from model list to entity list: %w", err)
	}

	// Bulk insert the SMA
	_, err = a.db.NamedExecContext(
		ctx,
		`INSERT INTO sma (exchange, pair, period, period_number, price_type, time, data)
		VALUES (:exchange, :pair, :period, :period_number, :price_type, :time, :data)
		ON CONFLICT (exchange, pair, period, period_number, price_type, time) DO UPDATE
		SET data = EXCLUDED.data`,
		entities.FromEntitiesToMap(ents),
	)
	if err != nil {
		return db.UpsertSMAActivityResults{}, fmt.Errorf("bulk inserting sma: %w", err)
	}

	return db.UpsertSMAActivityResults{}, nil
}
