package touch

import (
	"fmt"
	"slices"
	"strings"

	"go.viam.com/rdk/motionplan"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/rdk/spatialmath"
)

func HashString(s string) int {
	hash := 0
	for idx, c := range s {
		hash += ((idx + 1) * 7) + ((int(c) + 12) * 12)
	}
	return hash
}

func HashTransforms(transforms []*referenceframe.LinkInFrame) int {
	hash := 0

	slices.SortFunc(transforms, func(a, b *referenceframe.LinkInFrame) int {
		return strings.Compare(a.Name(), b.Name())
	})

	for _, l := range transforms {
		hash += HashLinkInFrame(l)
	}

	return hash
}

func HashLinkInFrame(lif *referenceframe.LinkInFrame) int {
	hash := HashPoseInFrame(lif.PoseInFrame)

	hash += HashString(fmt.Sprintf("%v", lif.Geometry())) // TODO hack

	return hash
}

func HashPoseInFrame(pif *referenceframe.PoseInFrame) int {

	hash := 0
	hash += HashString(pif.Name())
	hash += HashString(pif.Parent())

	hash += HashPose(pif.Pose())

	return hash
}

func HashPose(p spatialmath.Pose) int {
	hash := 0

	pp := p.Point()

	hash += (5 * (int(pp.X*10) + 100)) * 2
	hash += (6 * (int(pp.Y*10) + 10221)) * 3
	hash += (7 * (int(pp.Z*10) + 2124)) * 4

	o := p.Orientation().OrientationVectorDegrees()
	hash += (8 * (int(o.OX*100) + 2313)) * 5
	hash += (9 * (int(o.OY*100) + 3133)) * 6
	hash += (10 * (int(o.OZ*100) + 2931)) * 7
	hash += (11 * (int(o.Theta*10) + 6315)) * 8

	return hash
}

func HashMoveReq(req motion.MoveReq) int {
	hash := HashString(req.ComponentName.ShortName())
	hash += HashPoseInFrame(req.Destination)
	hash += HashConstraints(req.Constraints)
	hash += HashWorldState(req.WorldState)
	return hash
}

func HashConstraints(c *motionplan.Constraints) int {
	if c == nil {
		return 0
	}

	hash := 1

	for _, cc := range c.LinearConstraint {
		hash += HashLinearConstraint(cc)
	}
	for _, cc := range c.PseudolinearConstraint {
		hash += HashPseudolinearConstraint(cc)
	}
	for _, cc := range c.OrientationConstraint {
		hash += HashOrientationConstraint(cc)
	}
	for _, cc := range c.CollisionSpecification {
		hash += HashCollisionSpecification(cc)
	}

	return hash
}
func HashLinearConstraint(c motionplan.LinearConstraint) int {
	return int(c.LineToleranceMm*12321) + int(c.OrientationToleranceDegs*831)
}

func HashPseudolinearConstraint(c motionplan.PseudolinearConstraint) int {
	return int(c.LineToleranceFactor*72321) + int(c.OrientationToleranceFactor*83231)
}

func HashOrientationConstraint(c motionplan.OrientationConstraint) int {
	return int(c.OrientationToleranceDegs * 84132)

}
func HashCollisionSpecification(c motionplan.CollisionSpecification) int {
	hash := 0

	for _, a := range c.Allows {
		hash += HashString(a.Frame1)*12 + HashString(a.Frame2)*931
	}

	return hash
}

func HashInputs(in []referenceframe.Input) int {
	hash := 0
	for i, v := range in {
		hash += ((i + 7) * int(v.Value*10)) * (i + 11)
	}
	return hash
}

func HashWorldState(ws *referenceframe.WorldState) int {
	if ws == nil {
		return 0
	}

	hash := 0

	for _, o := range ws.Obstacles() {
		hash += HashString(fmt.Sprintf("%v", o)) // TODO HACK
	}

	for _, t := range ws.Transforms() {
		hash += HashLinkInFrame(t)
	}

	return hash
}
