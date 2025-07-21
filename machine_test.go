package vmodutils

import (
	"fmt"
	"testing"

	"go.viam.com/test"
)

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
