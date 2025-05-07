package touch

import (
	"testing"

	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/rimage/transform"
	"go.viam.com/test"
)

func TestRSPixelToPoint(t *testing.T) {

	p := camera.Properties{
		IntrinsicParams:  &transform.PinholeCameraIntrinsics{Width: 1280, Height: 720, Fx: 906.0663452148438, Fy: 905.1234741210938, Ppx: 646.94970703125, Ppy: 374.4667663574219},
		DistortionParams: &transform.BrownConrady{RadialK1: 0, RadialK2: 0, RadialK3: 0, TangentialP1: 0, TangentialP2: 0},
	}

	x, y, z := 592.0, 499.0, 459.601175

	xx, yy, zz := PixelToPoint(p, x, y, z)
	test.That(t, xx, test.ShouldAlmostEqual, -27.87, .01)
	test.That(t, yy, test.ShouldAlmostEqual, 63.23, .01)
	test.That(t, zz, test.ShouldAlmostEqual, z)

	xxx, yyy, zzz := RSPixelToPoint(p, x, y, z)
	test.That(t, zzz, test.ShouldAlmostEqual, z)

	// TODO - i think this is wrong??

	test.That(t, xxx, test.ShouldAlmostEqual, -27.87, .01)
	test.That(t, yyy, test.ShouldAlmostEqual, 63.23, .01)

	//test.That(t, xxx, test.ShouldAlmostEqual, -37.042980)
	//test.That(t, yyy, test.ShouldAlmostEqual, 62.105444)

}
