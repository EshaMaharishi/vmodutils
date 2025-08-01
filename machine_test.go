package vmodutils

import (
	"fmt"
	"testing"

	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/utils"
	"go.viam.com/test"
)

func helperMachineConfig(componentNames, serviceNames, fragmentIDs []string) map[string]interface{} {
	components := []interface{}{}
	for _, c := range componentNames {
		components = append(components, map[string]interface{}{"name": c})
	}
	services := []interface{}{}
	for _, s := range serviceNames {
		services = append(services, map[string]interface{}{"name": s})
	}

	return map[string]interface{}{"components": components,
		"services":  services,
		"fragments": fragmentIDs}
}

// simple helper function for tests
func getAttrFromConfigForTests(machine map[string]interface{}, name string) interface{} {
	cs, ok := machine["components"].([]interface{})
	if !ok {
		return nil
	}
	services, ok := machine["services"].([]interface{})
	if ok {
		cs = append(cs, services...)
	}

	for _, cc := range cs {
		ccc, ok := cc.(map[string]interface{})
		if !ok {
			return nil
		}
		if ccc["name"] != name {
			continue
		}

		return ccc["attributes"]
	}
	return nil
}

func TestUpdateComponentOrServiceConfig(t *testing.T) {
	newAttr := utils.AttributeMap{"attr1": true}
	cases := []struct {
		description   string
		componentName string
		shouldBeFound bool
		expectedErr   error
		machineConfig map[string]interface{}
	}{
		{
			description:   "the component is in the machine",
			componentName: "c1",
			shouldBeFound: true,
			machineConfig: helperMachineConfig([]string{"c1", "c2"}, []string{"s1", "s2"}, []string{"f1", "f2"}),
		},
		{
			description:   "the service is in the machine",
			componentName: "s2",
			shouldBeFound: true,
			machineConfig: helperMachineConfig([]string{"c1", "c2"}, []string{"s1", "s2"}, []string{"f1", "f2"}),
		},
		{
			description:   "the component cannot be found",
			componentName: "c3",
			shouldBeFound: false,
			machineConfig: helperMachineConfig([]string{"c1", "c2"}, []string{"s1", "s2"}, []string{"f1", "f2"}),
		},
		{
			description:   "really bad machine config",
			componentName: "not-here",
			shouldBeFound: false,
			machineConfig: nil,
			expectedErr:   fmt.Errorf("no components %T", nil),
		},
	}
	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			// make a machine for the test
			name := resource.NewName(resource.APINamespaceRDK.WithComponentType("test"), tt.componentName)
			found, err := updateComponentOrServiceConfig(tt.machineConfig, name, newAttr)
			test.That(t, found, test.ShouldEqual, tt.shouldBeFound)
			test.That(t, err, test.ShouldResemble, tt.expectedErr)
			if tt.shouldBeFound {
				updatedAttrs := getAttrFromConfigForTests(tt.machineConfig, tt.componentName)
				test.That(t, updatedAttrs, test.ShouldResemble, newAttr)
			}
		})
	}
}
func TestCheckSameInputs(t *testing.T) {
	id := "id"
	version := "version"
	reallyBadFrag := 25
	cases := []struct {
		description     string
		frag            interface{}
		expectedID      string
		expectedVersion string
		expectedErr     error
	}{
		{
			description: "the fragment is just the id string",
			frag:        id,
			expectedID:  id,
		},
		{
			description: "fragment is a map with no version",
			frag:        map[string]interface{}{"id": id},
			expectedID:  id,
		},
		{
			description:     "fragment is a map with a version",
			frag:            map[string]interface{}{"id": id, "version": version},
			expectedID:      id,
			expectedVersion: version,
		},
		{
			description: "fragment has no id",
			frag:        map[string]interface{}{"version": version},
			expectedID:  "",
			expectedErr: fmt.Errorf("fragment is missing an id: %v", map[string]interface{}{"version": version}),
		},
		{
			description: "fragment is not an expected fragment config",
			frag:        reallyBadFrag,
			expectedID:  "",
			expectedErr: fmt.Errorf("fragment config does not match expected interface: %T", reallyBadFrag),
		},
	}
	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			idOut, versionOut, err := checkFragmentInConfig(tt.frag)
			test.That(t, idOut, test.ShouldEqual, tt.expectedID)
			test.That(t, versionOut, test.ShouldResemble, tt.expectedVersion)
			test.That(t, err, test.ShouldResemble, tt.expectedErr)

		})
	}
}
