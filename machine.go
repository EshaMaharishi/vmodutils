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
	cs, ok := part.RobotConfig["components"].([]interface{})
	if !ok {
		return fmt.Errorf("no components %T", part.RobotConfig["components"])
	}
	services, ok := part.RobotConfig["services"].([]interface{})
	if ok {
		cs = append(cs, services...)
	}
	fragments, hasFragments := part.RobotConfig["fragments"].([]interface{})

	found := false

	for idx, cc := range cs {
		ccc, ok := cc.(map[string]interface{})
		if !ok {
			return fmt.Errorf("config bad %d: %T", idx, cc)
		}
		if ccc["name"] != name.ShortName() {
			continue
		}

		ccc["attributes"] = newAttr
		found = true
	}

	// check fragments
	if !found && hasFragments {
		for idx, frag := range fragments {
			id, version, err := checkFragmentInConfig(frag)
			if err != nil {
				return err
			}
			fragModString := ""
			// first, determine which fragment has the component.
			found, fragModString, err = findComponentInFragment(ctx, c, id, version, name)
			if err != nil {
				continue
			}
			if found {
				// find fragment_mods. swallow the error because these are not required
				fragMods, _ := part.RobotConfig["fragment_mods"].([]interface{})
				foundSet := false

				for _, fragMod := range fragMods {
					fragModc, ok := fragMod.(map[string]interface{})
					if !ok {
						return fmt.Errorf("config bad %d: %T", idx, fragMod)
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
						sets, ok := modc["$set"].(map[string]interface{})
						if !ok {
							// there are no mods set for this fragment. break out to create one
							break
						}

						// check the keys to see if the set if for our component
						for k := range sets {
							if strings.Contains(k, fragModString) {
								foundSet = true
								break
							}
						}
						if !foundSet {
							continue
						}
						// we found our component, go ahead and replace the component's mods
						attrMod := attrMapToFragmentMod(fragModString, newAttr)
						mods[indexMods] = attrMod
						foundSet = true

						break
					}
					// we found mods but we did not find any for our component. add a new set of mods
					if !foundSet {
						fragModc["mods"] = append(mods, attrMapToFragmentMod(fragModString, newAttr))
						foundSet = true
					}

				}
				// if we did not find any mods for our fragment. so add everything
				if !foundSet {
					newFragMod := map[string]interface{}{"fragment_id": id}
					newMods := attrMapToFragmentMod(fragModString, newAttr)
					newFragMod["mods"] = []map[string]interface{}{newMods}
					fragMods = append(fragMods, newFragMod)
					part.RobotConfig["fragment_mods"] = fragMods
				}
				// stop looking at fragments
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

func findComponentInFragment(ctx context.Context, c *app.AppClient, id string, version string, name resource.Name) (bool, string, error) {
	frag, err := c.GetFragment(ctx, id, version)
	if err != nil {
		return false, "", err
	}

	// components might not be defined, so it is ok to swallow the error here
	cs, _ := frag.Fragment["components"].([]interface{})
	for idx, cc := range cs {
		ccc, ok := cc.(map[string]interface{})
		if !ok {
			return false, "", fmt.Errorf("config bad %d: %T", idx, cc)
		}
		if ccc["name"] != name.ShortName() {
			continue
		}
		// we found the component within this fragment, return the fragment mod string
		return true, fmt.Sprintf("components.%s.attributes", name.ShortName()), nil
	}
	// services might not be defined, so it is ok to swallow the error here
	services, _ := frag.Fragment["services"].([]interface{})
	for idx, sc := range services {
		scc, ok := sc.(map[string]interface{})
		if !ok {
			return false, "", fmt.Errorf("config bad %d: %T", idx, sc)
		}
		if scc["name"] != name.ShortName() {
			continue
		}
		// we found the service within this fragment, return the fragment mod string
		return true, fmt.Sprintf("services.%s.attributes", name.ShortName()), nil

	}

	// check fragments within fragments
	fragments, _ := frag.Fragment["fragments"].([]interface{})
	for _, fc := range fragments {
		idFrag, versionFrag, err := checkFragmentInConfig(fc)
		if err != nil {
			return false, "", err
		}

		found, fragModString, err := findComponentInFragment(ctx, c, idFrag, versionFrag, name)
		if err != nil {
			return false, "", err
		}
		if found {
			return true, fragModString, nil
		}
	}
	// we did not find the component in this fragment.
	return false, "", nil
}

func FindDep(deps resource.Dependencies, n string) (resource.Resource, bool) {
	for nn, r := range deps {
		if nn.ShortName() == n {
			return r, true
		}
	}
	return nil, false
}
