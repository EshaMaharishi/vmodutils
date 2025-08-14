package touch

import (
	"context"
	"fmt"

	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/robot"
	"go.viam.com/rdk/robot/framesystem"
)

func FrameSystemWithSomeParts(ctx context.Context, myRobot robot.Robot, names []string, transforms []*referenceframe.LinkInFrame) (*referenceframe.FrameSystem, error) {
	fsc, err := myRobot.FrameSystemConfig(ctx)
	if err != nil {
		return nil, err
	}

	parts := []*referenceframe.FrameSystemPart{}

	seen := map[string]bool{}

	for _, name := range names {
		for name != "world" {
			p := FindPart(fsc, name)
			if p == nil {
				return nil, fmt.Errorf("cannot find frame [%s]", name)
			}
			if !seen[name] {
				parts = append(parts, p)
				seen[name] = true
			}
			name = p.FrameConfig.Parent()
		}
	}

	return referenceframe.NewFrameSystem("temp", parts, transforms)
}

func FindPart(fsc *framesystem.Config, name string) *referenceframe.FrameSystemPart {
	for _, c := range fsc.Parts {
		if c.FrameConfig.Name() == name {
			return c
		}
	}
	return nil
}
