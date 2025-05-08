package touch

import (
	"image"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/golang/geo/r3"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/rimage"
	"go.viam.com/test"
)

func TestPCD1(t *testing.T) {
	logger := logging.NewTestLogger(t)

	pc, err := pointcloud.NewFromFile("data/test.pcd", logger)
	test.That(t, err, test.ShouldBeNil)

	cleaner, err := pointcloud.StatisticalOutlierFilter(50, 1)
	test.That(t, err, test.ShouldBeNil)

	pc, err = cleaner(pc)
	test.That(t, err, test.ShouldBeNil)

	b1 := PrepBoundingRectForSearch()
	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		if math.Abs(p.Z) < 10 && d.HasColor() {
			BoundingRectMinMax(b1, p)
		}
		return true
	})

	logger.Infof("b1 %v -> %v", b1, b1.Size())

	b2 := PrepBoundingRectForSearch()

	img := image.NewRGBA(image.Rectangle{Max: b1.Size()})
	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		if math.Abs(p.Z) < 25 && d.HasColor() {
			x := int(p.X)
			y := int(p.Y)

			x = x - b1.Min.X
			y = y - b1.Min.Y

			car := math.Pow((p.X*p.X)+(p.Y*p.Y), .5)

			dis := rimage.White.Distance(rimage.NewColorFromColor(d.Color()))

			if dis < 3 && car > 100 {
				if x > 400 {
					logger.Infof("%v,%v -> %v,%v", int(p.X), int(p.Y), x, y)
				}
				img.Set(x, y, d.Color())
				BoundingRectMinMax(b2, p)

				//img.Set(x,y,rimage.White)
			}

		}

		return true
	})

	logger.Infof("b2 %v %v", b2, b2.Size())

	file, err := os.Create("test.png")
	test.That(t, err, test.ShouldBeNil)
	defer file.Close()

	err = png.Encode(file, img)
	test.That(t, err, test.ShouldBeNil)

	h := PCDFindHighestInRegion(pc, *b2)
	logger.Infof("hi %v", h)
}

func TestPCDCrop(t *testing.T) {
	a := pointcloud.New()
	a.Set(r3.Vector{1, 1, 1}, pointcloud.NewBasicData())
	a.Set(r3.Vector{5, 5, 5}, pointcloud.NewBasicData())
	a.Set(r3.Vector{9, 9, 9}, pointcloud.NewBasicData())
	a.Set(r3.Vector{5, 0, 5}, pointcloud.NewBasicData())

	test.That(t, a.Size(), test.ShouldEqual, 4)
	_, got := a.At(1, 1, 1)
	test.That(t, got, test.ShouldBeTrue)
	_, got = a.At(5, 5, 5)
	test.That(t, got, test.ShouldBeTrue)
	_, got = a.At(5, 0, 5)
	test.That(t, got, test.ShouldBeTrue)

	b := PCDCrop(a, r3.Vector{2, 2, 2}, r3.Vector{7, 7, 7})
	test.That(t, b.Size(), test.ShouldEqual, 1)
	_, got = b.At(1, 1, 1)
	test.That(t, got, test.ShouldBeFalse)
	_, got = b.At(5, 5, 5)
	test.That(t, got, test.ShouldBeTrue)

}

func TestPCDCup1(t *testing.T) {
	logger := logging.NewTestLogger(t)

	pc, err := pointcloud.NewRoundingPointCloudFromFile("data/cup1.pcd", logger)
	test.That(t, err, test.ShouldBeNil)

	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		logger.Infof("%v", p)
		return true
	})

	t.Fail()

}
