package touch

import (
	"context"
	"fmt"
	"sync"
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
	OtherFrames      []string `json:"other_frames"`
	Fallback         bool
	MaxJointDistance float64 `json:"max_joint_distance"`
}

func (cfg *SingleArmConfig) Validate(path string) ([]string, []string, error) {
	if cfg.Arm == "" {
		return nil, nil, fmt.Errorf("need an arm")
	}
	return []string{cfg.Arm, motion.Named("builtin").String()}, nil, nil
}

func (cfg *SingleArmConfig) maxJointDistance() float64 {
	if cfg.MaxJointDistance <= 0 {
		return 1.5
	}
	return cfg.MaxJointDistance
}

func (cfg *SingleArmConfig) allFrames() []string {
	all := []string{cfg.Arm}
	all = append(all, cfg.OtherFrames...)
	return all
}

type fsCacheEntry struct {
	fs referenceframe.FrameSystem

	plansLock sync.Mutex
	plans     map[int][][]referenceframe.Input
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

	cachedFSLock sync.Mutex
	cachedFS     map[int]*fsCacheEntry

	planHashLock sync.Mutex
	planHash     map[string][][]float64
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
		name:     name,
		logger:   logger,
		cfg:      conf,
		cachedFS: map[int]*fsCacheEntry{},
		planHash: map[string][][]float64{},
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

	fs, err := FrameSystemWithSomeParts(ctx, s.robotClient, conf.allFrames(), nil)
	if err != nil {
		return nil, err
	}

	s.cachedFS[0] = &fsCacheEntry{fs: fs, plans: map[int][][]referenceframe.Input{}}

	return s, nil
}

func (s *singleArmService) Name() resource.Name {
	return s.name
}

func (s *singleArmService) getFS(ctx context.Context, req motion.MoveReq) (*fsCacheEntry, error) {
	h := 0

	if req.WorldState != nil && len(req.WorldState.Transforms()) > 0 {
		h = HashTransforms(req.WorldState.Transforms())
	}

	s.cachedFSLock.Lock()
	temp, ok := s.cachedFS[h]
	s.cachedFSLock.Unlock()

	if ok {
		s.logger.Debugf("using cached fs %v %v", h, req.WorldState.Transforms())
		return temp, nil
	}

	s.logger.Infof("creating new cached fs %v %v", h, req.WorldState.Transforms())

	fs, err := FrameSystemWithSomeParts(ctx, s.robotClient, s.cfg.allFrames(), req.WorldState.Transforms())
	if err != nil {
		return nil, err
	}

	temp = &fsCacheEntry{fs: fs, plans: map[int][][]referenceframe.Input{}}

	s.cachedFSLock.Lock()
	s.cachedFS[h] = temp
	s.cachedFSLock.Unlock()

	return temp, nil
}

func (s *singleArmService) getPlan(ctx context.Context, req motion.MoveReq, fs *fsCacheEntry, startJoints []referenceframe.Input) ([][]referenceframe.Input, error) {

	planHash := HashMoveReq(req) + (3 * HashInputs(startJoints))

	fs.plansLock.Lock()
	myPlan, ok := fs.plans[planHash]
	fs.plansLock.Unlock()

	s.logger.Infof("getPlan \n\t hash: %v ok: %v \n\t start: %v \n\t dest: %v", planHash, ok, startJoints, req.Destination)

	if ok {
		s.logger.Infof("using cached plan %v", planHash)
		return myPlan, nil
	}

	myPlan, err := s.createPlan(ctx, req, fs.fs, startJoints)
	if err != nil {
		return nil, err
	}

	fs.plansLock.Lock()
	fs.plans[planHash] = myPlan
	fs.plansLock.Unlock()

	return myPlan, nil
}

func (s *singleArmService) createPlan(ctx context.Context, req motion.MoveReq, myFs referenceframe.FrameSystem, startJoints []referenceframe.Input) ([][]referenceframe.Input, error) {

	planReq := &motionplan.PlanRequest{
		Logger:      s.logger,
		FrameSystem: myFs,
		Goals: []*motionplan.PlanState{
			motionplan.NewPlanState(referenceframe.FrameSystemPoses{req.ComponentName.ShortName(): req.Destination}, nil),
		},
		StartState: motionplan.NewPlanState(nil, referenceframe.FrameSystemInputs{s.cfg.Arm: startJoints}),
		WorldState: req.WorldState,
	}

	startTime := time.Now()
	plan, err := motionplan.PlanMotion(ctx, planReq)
	if err != nil {
		return nil, err
	}

	s.logger.Infof("plan: trajectory length: %d path length: %d, planned in %v", len(plan.Trajectory()), len(plan.Path()), time.Since(startTime))

	myPlan := [][]referenceframe.Input{}

	prev := startJoints
	for _, t := range plan.Trajectory() {
		myPlan = append(myPlan, t[s.cfg.Arm])
		distance := referenceframe.InputsL2Distance(prev, t[s.cfg.Arm])
		if distance > s.cfg.maxJointDistance() {
			return nil, fmt.Errorf("distance: %v > maxJointDistance (%v)", distance, s.cfg.maxJointDistance())
		}
	}

	return myPlan, nil
}

func (s *singleArmService) Move(ctx context.Context, req motion.MoveReq) (bool, error) {

	myFs, err := s.getFS(ctx, req)
	if err != nil {
		return false, err
	}

	startJoints, err := s.myArm.JointPositions(ctx, nil)
	if err != nil {
		return false, err
	}

	myPlan, err := s.getPlan(ctx, req, myFs, startJoints)
	if err != nil {
		if s.cfg.Fallback {
			s.logger.Info("falling back because of %v", err)
			return s.builtinMotion.Move(ctx, req)
		}
		return false, err
	}

	if req.Extra != nil && req.Extra["hash"] == true {
		hashKey, ok := req.Extra["hash_key"].(string)
		if !ok {
			return false, fmt.Errorf("hash_key not a string")
		}

		newPlan := [][]float64{}

		for _, i := range myPlan {
			newPlan = append(newPlan, referenceframe.InputsToFloats(i))
		}

		s.planHashLock.Lock()
		s.planHash[hashKey] = newPlan
		s.planHashLock.Unlock()

		return false, nil
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
	if cmd["get_hash"] == true {
		hashKey, ok := cmd["hash_key"].(string)
		if !ok {
			return nil, fmt.Errorf("hash_key not a string")
		}

		s.planHashLock.Lock()
		plan, ok := s.planHash[hashKey]
		s.planHashLock.Unlock()

		if !ok {
			return nil, fmt.Errorf("no hash entry with %v", plan)
		}

		return map[string]interface{}{"plan": plan}, nil

	}
	return nil, fmt.Errorf("unknown command %v", cmd)
}

func (s *singleArmService) Close(ctx context.Context) error {
	return s.robotClient.Close(ctx)
}
