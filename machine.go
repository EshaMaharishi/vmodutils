package vmodutils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.viam.com/rdk/app"
	"go.viam.com/rdk/cli"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/robot"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
)

func MachineToDependencies(client robot.Robot) (resource.Dependencies, error) {
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
	params := []string{}
	for _, pp := range []string{utils.MachineFQDNEnvVar, utils.APIKeyIDEnvVar, utils.APIKeyEnvVar} {
		x := os.Getenv(pp)
		if x == "" {
			return nil, fmt.Errorf("no environment variable for %s", pp)
		}
		params = append(params, x)
	}
	return ConnectToMachine(ctx, logger, params[0], params[1], params[2])
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

// ConnectToHostFromCLIToken uses the viam cli token to login to a machine with just a hostname.
// use "viam login" to setup the token.
func ConnectToHostFromCLIToken(ctx context.Context, host string, logger logging.Logger) (robot.Robot, error) {
	if host == "" {
		return nil, fmt.Errorf("need to specify host")
	}

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

func UpdateComponentCloudAttributesFromModuleEnv(ctx context.Context, name resource.Name, newAttr utils.AttributeMap, logger logging.Logger) error {
	id := os.Getenv(utils.MachinePartIDEnvVar)
	if id == "" {
		return fmt.Errorf("no %s in env", utils.MachinePartIDEnvVar)
	}

	c, err := app.CreateViamClientFromEnvVars(ctx, nil, logger)
	if err != nil {
		return err
	}
	defer c.Close()

	return UpdateComponentCloudAttributes(ctx, c.AppClient(), id, name, newAttr)

}

func UpdateComponentCloudAttributes(ctx context.Context, c *app.AppClient, id string, name resource.Name, newAttr utils.AttributeMap) error {
	part, _, err := c.GetRobotPart(ctx, id)
	if err != nil {
		return err
	}
	found, err := updateComponentOrServiceConfig(part.RobotConfig, name, newAttr)
	if err != nil {
		return err
	}

	fragments, hasFragments := part.RobotConfig["fragments"].([]interface{})

	// check fragments
	if !found && hasFragments {
		for _, frag := range fragments {
			fID, version, err := checkFragmentInConfig(frag)
			if err != nil {
				return err
			}
			// first, determine which fragment has the component.
			fragModString, err := findComponentInFragment(ctx, c, fID, version, name)
			if err != nil {
				continue
			}
			if fragModString != "" {
				updateFragmentConfig(fID, fragModString, part.RobotConfig, attrMapToFragmentMod(fragModString, newAttr))
				found = true
				break
			}
		}

	}
	if !found {
		return fmt.Errorf("didn't find component with name %v", name.ShortName())
	}

	_, err = c.UpdateRobotPart(ctx, id, part.Name, part.RobotConfig)
	return err
}

func updateComponentOrServiceConfig(robotConfig map[string]interface{}, name resource.Name, newAttr utils.AttributeMap) (bool, error) {
	cs, ok := robotConfig["components"].([]interface{})
	if !ok {
		return false, fmt.Errorf("no components %T", robotConfig["components"])
	}
	services, ok := robotConfig["services"].([]interface{})
	if ok {
		cs = append(cs, services...)
	}

	found := false

	for idx, cc := range cs {
		ccc, ok := cc.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("config bad %d: %T", idx, cc)
		}
		if ccc["name"] != name.ShortName() {
			continue
		}

		ccc["attributes"] = newAttr
		found = true
	}
	return found, nil
}

func updateFragmentConfig(id, fragModString string, robotConfig, fragmentMod map[string]interface{}) error {
	fragMods, _ := robotConfig["fragment_mods"].([]interface{})
	for _, fragMod := range fragMods {
		fragModc, ok := fragMod.(map[string]interface{})
		if !ok {
			return fmt.Errorf("fragment mod config bad for fragment %v: %T", id, fragMod)
		}
		// check if we found the fragment that we want to modify
		if fragModc["fragment_id"] != id {
			continue
		}
		// find the mods for the fragment. app will strip out a defined mod that is empty so we do not have to check for that
		mods, ok := fragModc["mods"].([]interface{})
		if !ok {
			// we did not find mods for the fragment, break the loop and create our own
			break
		}
		//
		for indexMods, mod := range mods {
			modc, _ := mod.(map[string]interface{})
			sets, _ := modc["$set"].(map[string]interface{})
			// check the keys to see if the set if for our component
			for k := range sets {
				if strings.Contains(k, fragModString) {
					// we found our component, go ahead and replace the component's mods
					mods[indexMods] = fragmentMod
					return nil
				}
			}

		}
		// we found mods but we did not find any for our component. add a new set of mods
		fragModc["mods"] = append(mods, fragmentMod)
		return nil
	}
	// we don't have any fragment mods, so add a new set of mods to the config
	newFragMod := map[string]interface{}{"fragment_id": id}
	newFragMod["mods"] = []map[string]interface{}{fragmentMod}
	fragMods = append(fragMods, newFragMod)
	robotConfig["fragment_mods"] = fragMods
	return nil
}

func attrMapToFragmentMod(fragModString string, newAttr utils.AttributeMap) map[string]interface{} {
	fragMods := map[string]interface{}{}
	mods := map[string]interface{}{}
	for key, value := range newAttr {
		mods[fmt.Sprintf("%s.%s", fragModString, key)] = value
	}
	fragMods["$set"] = mods
	return fragMods
}

// fragments can either be strings or a map[string]interface{}, so we need to check for both.
func checkFragmentInConfig(frag interface{}) (string, string, error) {
	// check if we are just an id
	id, ok := frag.(string)
	if ok {
		return id, "", nil
	}
	// check if we are a map[string]interface{}
	fragc, ok := frag.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("fragment config does not match expected interface: %T", frag)
	}

	// fragments have to have ids, and for some reason we have a fragment without one
	if id, ok = fragc["id"].(string); !ok {
		return "", "", fmt.Errorf("fragment is missing an id: %v", frag)
	}
	if version, ok := fragc["version"].(string); ok {
		return id, version, nil
	}
	return id, "", nil
}

func findComponentInFragment(ctx context.Context, c *app.AppClient, id string, version string, name resource.Name) (string, error) {
	frag, err := c.GetFragment(ctx, id, version)
	if err != nil {
		return "", err
	}

	// components might not be defined, so it is ok to swallow the error here
	cs, _ := frag.Fragment["components"].([]interface{})
	for idx, cc := range cs {
		ccc, ok := cc.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("config bad %d: %T", idx, cc)
		}
		if ccc["name"] != name.ShortName() {
			continue
		}
		// we found the component within this fragment, return the fragment mod string
		return fmt.Sprintf("components.%s.attributes", name.ShortName()), nil
	}
	// services might not be defined, so it is ok to swallow the error here
	services, _ := frag.Fragment["services"].([]interface{})
	for idx, sc := range services {
		scc, ok := sc.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("config bad %d: %T", idx, sc)
		}
		if scc["name"] != name.ShortName() {
			continue
		}
		// we found the service within this fragment, return the fragment mod string
		return fmt.Sprintf("services.%s.attributes", name.ShortName()), nil

	}

	// check fragments within fragments
	fragments, _ := frag.Fragment["fragments"].([]interface{})
	for _, fc := range fragments {
		idFrag, versionFrag, err := checkFragmentInConfig(fc)
		if err != nil {
			return "", err
		}

		fragModString, err := findComponentInFragment(ctx, c, idFrag, versionFrag, name)
		if err != nil {
			return "", err
		}
		if fragModString != "" {
			return fragModString, nil
		}
	}
	// we did not find the component in this fragment.
	return "", nil
}

func FindDep(deps resource.Dependencies, n string) (resource.Resource, bool) {
	for nn, r := range deps {
		if nn.ShortName() == n {
			return r, true
		}
	}
	return nil, false
}
