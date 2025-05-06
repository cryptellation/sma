//go:build integration
// +build integration

package sql

import (
	"context"
	"testing"

	"github.com/cryptellation/dbmigrator"
	"github.com/cryptellation/sma/configs"
	"github.com/cryptellation/sma/configs/sql/down"
	"github.com/cryptellation/sma/configs/sql/up"
	"github.com/cryptellation/sma/svc/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

func TestIndicatorsSuite(t *testing.T) {
	suite.Run(t, new(IndicatorsSuite))
}

type IndicatorsSuite struct {
	db.IndicatorsSuite
}

func (suite *IndicatorsSuite) SetupSuite() {
	act, err := New(context.Background(), viper.GetString(configs.EnvSQLDSN))
	suite.Require().NoError(err)

	mig, err := dbmigrator.NewMigrator(context.Background(), act.db, up.Migrations, down.Migrations, nil)
	suite.Require().NoError(err)
	suite.Require().NoError(mig.MigrateToLatest(context.Background()))

	suite.DB = act
}

func (suite *IndicatorsSuite) SetupTest() {
	db := suite.DB.(*Activities)
	suite.Require().NoError(db.Reset(context.Background()))
}
