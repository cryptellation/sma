//go:build e2e
// +build e2e

package test

import (
	"context"
)

func (suite *EndToEndSuite) TestServiceInfoWorkflow() {
	// WHEN requesting the service info

	info, err := suite.client.Info(context.Background())

	// THEN the request is successful

	suite.Require().NoError(err)

	// AND the response contains the proper version

	suite.Require().NotEqual("", info.Version)
}
