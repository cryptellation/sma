//go:build e2e
// +build e2e

package test

import (
	"context"
	"time"

	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/candlesticks/pkg/period"
	"github.com/cryptellation/sma/api"
)

func (suite *EndToEndSuite) TestListIndicators() {
	// WHEN requesting for SMA

	t1, _ := time.Parse(time.RFC3339, "2023-02-26T12:00:00Z")
	t2, _ := time.Parse(time.RFC3339, "2023-02-26T12:01:00Z")
	t3, _ := time.Parse(time.RFC3339, "2023-02-26T12:02:00Z")
	ts, err := suite.client.List(context.Background(), api.ListWorkflowParams{
		Exchange:     "binance",
		Pair:         "ETH-USDT",
		Period:       period.M1,
		Start:        t1,
		End:          t3,
		PeriodNumber: 3,
		PriceType:    candlestick.PriceTypeIsClose,
	})

	// THEN there is no error

	suite.Require().NoError(err)

	// AND the response contains the proper SMA

	suite.Require().Equal(3, ts.Data.Len())

	v, exists := ts.Data.Get(t1)
	suite.Require().True(exists)
	suite.Require().Equal(1603.8966666666668, v)

	v, exists = ts.Data.Get(t2)
	suite.Require().True(exists)
	suite.Require().Equal(1604.17, v)

	v, exists = ts.Data.Get(t3)
	suite.Require().True(exists)
	suite.Require().Equal(1604.3533333333335, v)
}
