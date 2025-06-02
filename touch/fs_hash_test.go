package touch

import (
	"testing"

	"github.com/golang/geo/r3"

	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/test"
)

func TestHashString(t *testing.T) {
	test.That(t, HashString(""), test.ShouldEqual, HashString(""))
	test.That(t, HashString("eliot"), test.ShouldEqual, HashString("eliot"))

	test.That(t, HashString(""), test.ShouldNotEqual, HashString("a"))
	test.That(t, HashString(""), test.ShouldNotEqual, HashString("eliot"))
}

func TestHashInputs(t *testing.T) {
	a := []referenceframe.Input{{5}, {7}}
	b := []referenceframe.Input{{5.0001}, {7.0001}}
	c := []referenceframe.Input{{5.1}, {7.1}}
	d := []referenceframe.Input{{7}, {5}}

	test.That(t, HashInputs(a), test.ShouldEqual, HashInputs(a))
	test.That(t, HashInputs(a), test.ShouldEqual, HashInputs(b))
	test.That(t, HashInputs(a), test.ShouldNotEqual, HashInputs(c))
	test.That(t, HashInputs(a), test.ShouldNotEqual, HashInputs(d))
}

func TestHashPose(t *testing.T) {
	pp := r3.Vector{X: 191.423448, Y: 297.871741, Z: 365.630400}

	oa := &spatialmath.OrientationVectorDegrees{OX: 0.739191, OY: -0.551458, OZ: 0.386640, Theta: 101.893958}
	ob := &spatialmath.OrientationVectorDegrees{OX: 0.739191, OY: -0.551458, OZ: 0.364362, Theta: 101.893958}

	test.That(t, HashPose(spatialmath.NewPose(pp, oa)), test.ShouldNotEqual, HashPose(spatialmath.NewPose(pp, ob)))
}
