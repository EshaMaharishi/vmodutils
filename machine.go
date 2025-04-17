package vmodutils

import (
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/robot/client"
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
