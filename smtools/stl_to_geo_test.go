package smtools

import (
	"testing"

	"go.viam.com/test"
)

func TestSTLFileToGeometry1(t *testing.T) {
	g, err := STLFileToGeometry("data/forearm.stl")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, g, test.ShouldNotBeNil)

}
