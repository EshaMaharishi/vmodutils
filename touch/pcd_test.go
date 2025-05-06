package touch

import (
	"fmt"
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

	fmt.Printf("b1 %v -> %v\n", b1, b1.Size())

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
					fmt.Printf("%v,%v -> %v,%v\n", int(p.X), int(p.Y), x, y)
				}
				img.Set(x, y, d.Color())
				BoundingRectMinMax(b2, p)

				//img.Set(x,y,rimage.White)
			}

		}

		return true
	})

	fmt.Printf("b2 %v %v\n", b2, b2.Size())

	file, err := os.Create("test.png")
	test.That(t, err, test.ShouldBeNil)
	defer file.Close()

	err = png.Encode(file, img)
	test.That(t, err, test.ShouldBeNil)

	h := PCDFindHighestInRegion(pc, *b2)
	fmt.Printf("hi %v\n", h)

}
