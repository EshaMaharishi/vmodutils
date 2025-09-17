package main

import (
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/components/gripper"
	"go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"

	"github.com/erh/vmodutils/touch"
)

func main() {
	module.ModularMain(
		resource.APIModel{camera.API, touch.CropCameraModel},
		resource.APIModel{camera.API, touch.MergeModel},
		resource.APIModel{camera.API, touch.MultipleArmPosesModel},
		resource.APIModel{toggleswitch.API, touch.ArmPositionSaverModel},
		resource.APIModel{gripper.API, touch.ObstacleModel},
		resource.APIModel{gripper.API, touch.ObstacleOpenBoxModel},
	)

}
