package main

import (
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"

	"github.com/erh/vmodutils/touch"
)

func main() {
	module.ModularMain(
		resource.APIModel{camera.API, touch.CropCameraModel},
	)

}
