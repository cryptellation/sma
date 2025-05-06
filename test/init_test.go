//go:build e2e
// +build e2e

package test

import (
	"testing"

	"github.com/cryptellation/sma/configs"
	"github.com/cryptellation/sma/pkg/clients"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	temporalclient "go.temporal.io/sdk/client"
)

func TestEndToEndSuite(t *testing.T) {
	suite.Run(t, new(EndToEndSuite))
}

type EndToEndSuite struct {
	suite.Suite
	client         clients.Client
	temporalclient temporalclient.Client
}

func (suite *EndToEndSuite) SetupSuite() {
	tc, err := temporalclient.Dial(temporalclient.Options{
		HostPort: viper.GetString(configs.EnvTemporalAddress),
	})
	suite.Require().NoError(err)
	suite.temporalclient = tc

	suite.client = clients.New(tc)
}

func (suite *EndToEndSuite) TearDownSuite() {
	suite.temporalclient.Close()
}
