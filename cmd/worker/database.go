package main

import (
	"context"
	"strconv"

	"github.com/cenkalti/backoff/v5"
	"github.com/cryptellation/dbmigrator"
	"github.com/cryptellation/sma/configs"
	"github.com/cryptellation/sma/configs/sql/down"
	"github.com/cryptellation/sma/configs/sql/up"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	driverNameFlag string
	dsnFlag        string
)

var (
	db *sqlx.DB
)

var databaseCmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{"i"},
	Short:   "Manage database",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) (err error) {
		db, err = loadDB(cmd.Context())
		return err
	},
}

func loadDB(ctx context.Context) (*sqlx.DB, error) {
	// Set backoff callback
	callback := func() (*sqlx.DB, error) {
		return sqlx.ConnectContext(ctx, driverNameFlag, dsnFlag)
	}

	// Retry with backoff
	return backoff.Retry(ctx, callback,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(10))
}

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"m"},
	Short:   "Migrate the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a migrator client
		mig, err := dbmigrator.NewMigrator(cmd.Context(), db, up.Migrations, down.Migrations, nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return mig.MigrateToLatest(cmd.Context())
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		return mig.MigrateTo(cmd.Context(), id)
	},
}

var rollbackCmd = &cobra.Command{
	Use:     "rollback",
	Aliases: []string{"r"},
	Short:   "Rollback the databas before a migration ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a migrator client
		mig, err := dbmigrator.NewMigrator(cmd.Context(), db, up.Migrations, down.Migrations, nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return mig.Rollback(cmd.Context())
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		return mig.RollbackUntil(cmd.Context(), id)
	},
}

func addDatabaseCommands(cmd *cobra.Command) {
	databaseCmd.AddCommand(migrateCmd)
	databaseCmd.AddCommand(rollbackCmd)

	// Set flags
	dsn := viper.GetString(configs.EnvSQLDSN)
	databaseCmd.PersistentFlags().StringVarP(&driverNameFlag, "driver", "d", "postgres", "Set the database driver name")
	databaseCmd.PersistentFlags().StringVarP(&dsnFlag, "dsn", "s", dsn, "Set the database data source name")

	cmd.AddCommand(databaseCmd)
}
