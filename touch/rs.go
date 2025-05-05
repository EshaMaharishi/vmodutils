package touch

import (
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/rimage/transform"
)

func PixelToPoint(properties camera.Properties, pixelX, pixelY, depth float64) (float64, float64, float64) {
	x, y, z := properties.IntrinsicParams.PixelToPoint(pixelX, pixelY, depth)

	if properties.DistortionParams != nil {
		x, y = properties.DistortionParams.Transform(x, y)
	}

	return x, y, z
}

func RSPixelToPoint(properties camera.Properties, pixelX, pixelY, depth float64) (float64, float64, float64) {
	intrin := properties.IntrinsicParams

	x := (pixelX - intrin.Ppx) / intrin.Fx
	y := (pixelY - intrin.Ppy) / intrin.Fy

	xo := x
	yo := y

	brown, ok := properties.DistortionParams.(*transform.BrownConrady)
	if ok {
		for i := 0; i < 10; i++ {
			r2 := x*x + y*y
			icdist := 1 / (1 + ((brown.TangentialP2*r2+brown.RadialK2)*r2+brown.RadialK1)*r2)
			delta_x := 2*brown.RadialK3*x*y + brown.TangentialP1*(r2+2*x*x)
			delta_y := 2*brown.TangentialP1*x*y + brown.RadialK3*(r2+2*y*y)
			x = (xo - delta_x) * icdist
			y = (yo - delta_y) * icdist
		}
	} else if properties.DistortionParams != nil {
		x, y = properties.DistortionParams.Transform(x, y)
	}

	return depth * x, depth * y, depth
}
