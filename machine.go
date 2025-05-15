package vmodutils

import (
	"context"
	"os"

	"go.viam.com/rdk/cli"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/robot"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
)

func MachineToDependencies(client *client.RobotClient) (resource.Dependencies, error) {
	deps := resource.Dependencies{}

	names := client.ResourceNames()
	for _, n := range names {
		r, err := client.ResourceByName(n)
		if err != nil {
			return nil, err
		}
		deps[n] = r
	}

	return deps, nil
}

func ConnectToMachineFromEnv(ctx context.Context, logger logging.Logger) (robot.Robot, error) {
	host := os.Getenv(utils.MachineFQDNEnvVar)
	apiKeyId := os.Getenv(utils.APIKeyIDEnvVar)
	apiKey := os.Getenv(utils.APIKeyEnvVar)
	return ConnectToMachine(ctx, logger, host, apiKeyId, apiKey)
}

func ConnectToMachine(ctx context.Context, logger logging.Logger, host, apiKeyId, apiKey string) (robot.Robot, error) {
	return client.New(
		ctx,
		host,
		logger,
		client.WithDialOptions(rpc.WithEntityCredentials(
			apiKeyId,
			rpc.Credentials{
				Type:    rpc.CredentialsTypeAPIKey,
				Payload: apiKey,
			},
		)),
	)
}

func ConnectToHostFromCLIToken(ctx context.Context, host string, logger logging.Logger) (robot.Robot, error) {
	c, err := cli.ConfigFromCache(nil)
	if err != nil {
		return nil, err
	}

	dopts, err := c.DialOptions()
	if err != nil {
		return nil, err
	}

	return client.New(
		ctx,
		host,
		logger,
		client.WithDialOptions(dopts...),
	)
}
