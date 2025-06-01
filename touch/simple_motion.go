package touch

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/motionplan"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/robot"
	"go.viam.com/rdk/services/motion"

	"github.com/erh/vmodutils"
)

var SingleArmModel = vmodutils.NamespaceFamily.WithModel("single-arm-motion-service")

func init() {
	resource.RegisterService(motion.API, SingleArmModel,
		resource.Registration[motion.Service, *SingleArmConfig]{
			Constructor: newSingleArmMotionService,
		},
	)
}

type SingleArmConfig struct {
	Arm              string
	Fallback         bool
	MaxJointDistance float64 `json:"max_joint_distance"`
}

func (cfg *SingleArmConfig) Validate(path string) ([]string, []string, error) {
	if cfg.Arm == "" {
		return nil, nil, fmt.Errorf("need an arm")
	}
	return []string{cfg.Arm, "builtin"}, nil, nil
}

func (cfg *SingleArmConfig) maxJointDistance() float64 {
	if cfg.MaxJointDistance <= 0 {
		return 1.5
	}
	return cfg.MaxJointDistance
}

type singleArmService struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *SingleArmConfig

	robotClient robot.Robot
	fs          referenceframe.FrameSystem

	myArm         arm.Arm
	builtinMotion motion.Service
}

func newSingleArmMotionService(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (motion.Service, error) {
	conf, err := resource.NativeConfig[*SingleArmConfig](rawConf)
	if err != nil {
		return nil, err
	}

	return NewSingleArmService(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewSingleArmService(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *SingleArmConfig, logger logging.Logger) (motion.Service, error) {

	s := &singleArmService{
		name:   name,
		logger: logger,
		cfg:    conf,
	}

	var err error

	s.myArm, err = arm.FromDependencies(deps, conf.Arm)
	if err != nil {
		return nil, err
	}

	s.builtinMotion, err = motion.FromDependencies(deps, "builtin")
	if err != nil {
		return nil, err
	}

	s.robotClient, err = vmodutils.ConnectToMachineFromEnv(ctx, logger)
	if err != nil {
		return nil, err
	}

	s.fs, err = FrameSystemWithOnePart(ctx, s.robotClient, conf.Arm)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *singleArmService) Name() resource.Name {
	return s.name
}

func (s *singleArmService) Move(ctx context.Context, req motion.MoveReq) (bool, error) {

	startJoints, err := s.myArm.JointPositions(ctx, nil)
	if err != nil {
		return false, err
	}

	planReq := &motionplan.PlanRequest{
		Logger:      s.logger,
		FrameSystem: s.fs,
		Goals: []*motionplan.PlanState{
			motionplan.NewPlanState(referenceframe.FrameSystemPoses{s.cfg.Arm: req.Destination}, nil),
		},
		StartState: motionplan.NewPlanState(nil, referenceframe.FrameSystemInputs{s.cfg.Arm: startJoints}),
		WorldState: req.WorldState,
		Options:    req.Extra,
	}

	startTime := time.Now()
	plan, err := motionplan.PlanMotion(ctx, planReq)
	if err != nil {
		return false, err
	}

	s.logger.Infof("plan: trajectory length: %d path length: %d, planned in %v", len(plan.Trajectory()), len(plan.Path()), time.Since(startTime))

	myPlan := [][]referenceframe.Input{}

	prev := startJoints
	for _, t := range plan.Trajectory() {
		myPlan = append(myPlan, t[s.cfg.Arm])
		distance := referenceframe.InputsL2Distance(prev, t[s.cfg.Arm])
		if distance > s.cfg.maxJointDistance() {
			s.logger.Infof("\t distance: %v > maxJointDistance (%v)", distance, s.cfg.maxJointDistance())
			if s.cfg.Fallback {
				s.logger.Info("falling back")
				return s.builtinMotion.Move(ctx, req)
			}
			return false, fmt.Errorf("distance: %v > maxJointDistance (%v)", distance, s.cfg.maxJointDistance())
		}
	}

	err = s.myArm.MoveThroughJointPositions(ctx, myPlan, nil, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *singleArmService) MoveOnMap(ctx context.Context, req motion.MoveOnMapReq) (motion.ExecutionID, error) {
	id := uuid.New()
	return id, fmt.Errorf("MoveOnMap not supported by %v", SingleArmModel)
}

func (s *singleArmService) MoveOnGlobe(ctx context.Context, req motion.MoveOnGlobeReq) (motion.ExecutionID, error) {
	id := uuid.New()
	return id, fmt.Errorf("MoveOnGlobe not supported by %v", SingleArmModel)
}

func (s *singleArmService) GetPose(ctx context.Context, componentName resource.Name, destinationFrame string, supplementalTransforms []*referenceframe.LinkInFrame, extra map[string]interface{}) (*referenceframe.PoseInFrame, error) {
	return s.builtinMotion.GetPose(ctx, componentName, destinationFrame, supplementalTransforms, extra)
}

func (s *singleArmService) StopPlan(ctx context.Context, req motion.StopPlanReq) error {
	return fmt.Errorf("StopPlan not supported by %v", SingleArmModel)
}

func (s *singleArmService) ListPlanStatuses(ctx context.Context, req motion.ListPlanStatusesReq) ([]motion.PlanStatusWithID, error) {
	return nil, fmt.Errorf("ListPlanStatuses not supported by %v", SingleArmModel)
}

func (s *singleArmService) PlanHistory(ctx context.Context, req motion.PlanHistoryReq) ([]motion.PlanWithStatus, error) {
	return nil, fmt.Errorf("PlanHistory not supported by %v", SingleArmModel)
}

func (s *singleArmService) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (s *singleArmService) Close(ctx context.Context) error {
	return s.robotClient.Close(ctx)
}
