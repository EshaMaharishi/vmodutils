package touch

import (
	"slices"
	"strings"

	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/services/motion"
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

	// TODO - geometry

	return hash
}

func HashPoseInFrame(pif *referenceframe.PoseInFrame) int {

	hash := 0
	hash += HashString(pif.Name())
	hash += HashString(pif.Parent())

	p := pif.Pose()

	pp := p.Point()

	hash += (5 * (int(pp.X*10) + 100)) * 2
	hash += (6 * (int(pp.Y*10) + 10221)) * 3
	hash += (7 * (int(pp.Z*10) + 2124)) * 4

	o := p.Orientation().OrientationVectorDegrees()
	hash += (8 * (int(o.OX*10) + 2313)) * 5
	hash += (9 * (int(o.OY*10) + 3133)) * 6
	hash += (10 * (int(o.OZ*10) + 2931)) * 7
	hash += (11 * (int(o.Theta*10) + 6315)) * 8

	return hash
}

func HashMoveReq(req motion.MoveReq) int {
	hash := HashString(req.ComponentName.ShortName())
	hash += HashPoseInFrame(req.Destination)
	return hash
}

func HashInputs(in []referenceframe.Input) int {
	hash := 0
	for i, v := range in {
		hash += ((i + 7) * int(v.Value*10)) * (i + 11)
	}
	return hash
}
