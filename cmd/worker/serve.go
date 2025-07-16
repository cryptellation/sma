package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/cenkalti/backoff/v5"
	"github.com/cryptellation/health"
	"github.com/cryptellation/sma/api"
	"github.com/cryptellation/sma/configs"
	"github.com/cryptellation/sma/svc"
	"github.com/cryptellation/sma/svc/db/sql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	temporalwk "go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"s"},
	Short:   "Launch the service",
	RunE:    serve,
}

func serve(cmd *cobra.Command, _ []string) error {
	// Set up context that cancels on SIGTERM or SIGINT
	sigCtx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Create errgroup and context
	eg, ctx := errgroup.WithContext(sigCtx)

	// Health server
	h, err := setupAndStartHealthServer(ctx, eg)
	if err != nil {
		return err
	}

	// Temporal worker
	w, workerCleanup, err := setupWorker(ctx, eg)
	if err != nil {
		return err
	}
	defer workerCleanup()

	// Service
	if err := setupService(ctx, w); err != nil {
		return err
	}

	// Signal health server is ready
	h.Ready(true)
	defer h.Ready(false)

	// Wait for everything to be finished
	err = eg.Wait()
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}

// setupAndStartHealthServer initializes and starts the health server,
// returning the health server instance.
func setupAndStartHealthServer(ctx context.Context, eg *errgroup.Group) (*health.Health, error) {
	// Create health server
	h, err := health.New(
		viper.GetString(configs.EnvHealthAddress),
	)
	if err != nil {
		return nil, err
	}

	// Add to errgroup
	eg.Go(func() error {
		return h.Serve(ctx)
	})

	return h, nil
}

// setupWorker creates the temporal client and worker, and returns the worker and a cleanup function.
func setupWorker(ctx context.Context, eg *errgroup.Group) (temporalwk.Worker, func(), error) {
	// Create temporal client
	temporalClient, err := createTemporalClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Create temporal worker and add to errgroup
	w := temporalwk.New(temporalClient, api.WorkerTaskQueueName, temporalwk.Options{})
	eg.Go(func() error {
		runErr := make(chan error, 1)
		go func() {
			runErr <- w.Run(nil)
		}()
		defer close(runErr)
		select {
		case <-ctx.Done():
			w.Stop()
			<-runErr
			return ctx.Err()
		case err := <-runErr:
			return err
		}
	})

	// Cleanup function
	cleanup := func() { temporalClient.Close() }

	return w, cleanup, nil
}

// setupService creates the db and service and registers them to the worker.
func setupService(ctx context.Context, w temporalwk.Worker) error {
	// Create db client
	db, err := createDBClient(ctx)
	if err != nil {
		return err
	}
	db.Register(w)

	// Create service
	service := svc.New(db)
	service.Register(w)

	return nil
}

func createTemporalClient(ctx context.Context) (client.Client, error) {
	// Set backoff callback
	callback := func() (client.Client, error) {
		return client.Dial(client.Options{
			HostPort: viper.GetString(configs.EnvTemporalAddress),
		})
	}

	// Retry with backoff
	return backoff.Retry(ctx, callback,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(10))
}

func createDBClient(ctx context.Context) (*sql.Activities, error) {
	// Set backoff callback with dummy return value
	callback := func() (*sql.Activities, error) {
		return sql.New(ctx, viper.GetString(configs.EnvSQLDSN))
	}

	// Retry with backoff
	return backoff.Retry(ctx, callback,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(10))
}
