package vmodutils

import (
	"context"
	"flag"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/utils/rpc"
)

type MachineSetup struct {
	machine, apiKey, apiKeyId string
}

func (ms *MachineSetup) Valid() bool {
	return ms.machine != ""
}

func (ms *MachineSetup) Connect(ctx context.Context, logger logging.Logger) (*client.RobotClient, error) {
	return client.New(
		ctx,
		ms.machine,
		logger,
		client.WithDialOptions(rpc.WithEntityCredentials(
			ms.apiKeyId,
			rpc.Credentials{
				Type:    rpc.CredentialsTypeAPIKey,
				Payload: ms.apiKey,
			})),
	)
}

func AddMachineFlags() *MachineSetup {
	ms := &MachineSetup{}
	flag.StringVar(&ms.machine, "machine", "", "")
	flag.StringVar(&ms.apiKey, "api-key", "", "")
	flag.StringVar(&ms.apiKeyId, "api-key-id", "", "")
	return ms
}
