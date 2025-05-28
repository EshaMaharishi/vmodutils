package touch

import (
	"context"
	"fmt"

	"github.com/golang/geo/r3"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/rdk/utils"

	"github.com/erh/vmodutils"
)

var ArmPositionSaverModel = vmodutils.NamespaceFamily.WithModel("arm-position-saver")

func init() {
	resource.RegisterComponent(
		toggleswitch.API,
		ArmPositionSaverModel,
		resource.Registration[toggleswitch.Switch, *ArmPositionSaverConfig]{
			Constructor: newArmPositionSaver,
		})
}

type ArmPositionSaverConfig struct {
	Arm         string
	Joints      []float64
	Motion      string
	Point       r3.Vector
	Orientation spatialmath.OrientationVectorDegrees
}

func (c *ArmPositionSaverConfig) Validate(path string) ([]string, []string, error) {
	if c.Arm == "" {
		return nil, nil, fmt.Errorf("no arm specificed")
	}

	deps := []string{c.Arm}

	if c.Motion != "" {
		if c.Motion == "builtin" {
			deps = append(deps, motion.Named("builtin").String())
		} else {
			deps = append(deps, c.Motion)
		}
	}

	return deps, nil, nil
}

func newArmPositionSaver(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (toggleswitch.Switch, error) {
	newConf, err := resource.NativeConfig[*ArmPositionSaverConfig](config)
	if err != nil {
		return nil, err
	}

	arm, err := arm.FromDependencies(deps, newConf.Arm)
	if err != nil {
		return nil, err
	}

	aps := &ArmPositionSaver{
		name:   config.ResourceName(),
		cfg:    newConf,
		logger: logger,
		arm:    arm,
	}

	if newConf.Motion != "" {
		aps.motion, err = motion.FromDependencies(deps, newConf.Motion)
		if err != nil {
			return nil, err
		}
	}

	return aps, nil
}

type ArmPositionSaver struct {
	resource.AlwaysRebuild
	resource.TriviallyCloseable

	name   resource.Name
	cfg    *ArmPositionSaverConfig
	logger logging.Logger

	arm    arm.Arm
	motion motion.Service
}

func (aps *ArmPositionSaver) Name() resource.Name {
	return aps.name
}

func (aps *ArmPositionSaver) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (aps *ArmPositionSaver) SetPosition(ctx context.Context, position uint32, extra map[string]interface{}) error {
	if position == 0 {
		return nil
	}

	if position == 1 {
		return aps.saveCurrentPosition(ctx)
	}

	if position == 2 {
		return aps.goToSavePosition(ctx)
	}

	return fmt.Errorf("bad position: %d", position)
}

func (aps *ArmPositionSaver) GetPosition(ctx context.Context, extra map[string]interface{}) (uint32, error) {
	return 0, nil
}

func (aps *ArmPositionSaver) GetNumberOfPositions(ctx context.Context, extra map[string]interface{}) (uint32, error) {
	return 3, nil
}

func (aps *ArmPositionSaver) saveCurrentPosition(ctx context.Context) error {
	newConfig := utils.AttributeMap{
		"arm": aps.cfg.Arm,
	}

	if aps.cfg.Motion == "" {
		inputs, err := aps.arm.JointPositions(ctx, nil)
		if err != nil {
			return err
		}

		newConfig["joints"] = referenceframe.InputsToFloats(inputs)
	} else {
		p, err := aps.motion.GetPose(ctx, aps.arm.Name(), "world", nil, nil)
		if err != nil {
			return err
		}
		newConfig["point"] = p.Pose().Point()
		newConfig["orientation"] = p.Pose().Orientation().OrientationVectorDegrees()
	}

	return vmodutils.UpdateComponentCloudAttributesFromModuleEnv(ctx, aps.name, newConfig, aps.logger)
}

func (aps *ArmPositionSaver) goToSavePosition(ctx context.Context) error {
	if len(aps.cfg.Joints) > 0 {
		return aps.arm.MoveToJointPositions(ctx, referenceframe.FloatsToInputs(aps.cfg.Joints), nil)
	}

	if aps.motion != nil {
		pif := referenceframe.NewPoseInFrame(
			"world",
			spatialmath.NewPose(aps.cfg.Point, &aps.cfg.Orientation),
		)

		done, err := aps.motion.Move(
			ctx,
			motion.MoveReq{
				ComponentName: resource.Name{Name: aps.cfg.Arm},
				Destination:   pif,
				WorldState:    nil,
			},
		)
		if err != nil {
			return err
		}
		if !done {
			return fmt.Errorf("move didn't finish")
		}
		return nil
	}

	return fmt.Errorf("need to configure where to go")
}
