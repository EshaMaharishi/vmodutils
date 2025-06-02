package touch

import (
	"context"
	"fmt"

	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/robot"
	"go.viam.com/rdk/robot/framesystem"
)

func FrameSystemWithOnePart(ctx context.Context, myRobot robot.Robot, name string, transforms []*referenceframe.LinkInFrame) (referenceframe.FrameSystem, error) {
	fsc, err := myRobot.FrameSystemConfig(ctx)
	if err != nil {
		return nil, err
	}

	parts := []*referenceframe.FrameSystemPart{}

	for name != "world" {
		p := FindPart(fsc, name)
		if p == nil {
			return nil, fmt.Errorf("cannot find frame [%s]", name)
		}
		parts = append(parts, p)
		name = p.FrameConfig.Parent()
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
