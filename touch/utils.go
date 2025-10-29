package touch

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/golang/geo/r3"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/components/camera"
	toggleswitch "go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/spatialmath"
)

func PCFindHighestInRegion(pc pointcloud.PointCloud, box image.Rectangle) r3.Vector {

	best := r3.Vector{Z: -100000}

	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		if p.Z > best.Z {
			if p.X >= float64(box.Min.X) && p.Y >= float64(box.Min.Y) {
				if p.X <= float64(box.Max.X) && p.Y <= float64(box.Max.Y) {
					best = p
				}
			}
		}

		return true
	})

	return best
}

func PCFindLowestInRegion(pc pointcloud.PointCloud, box image.Rectangle) r3.Vector {

	best := r3.Vector{Z: 100000}

	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		if p.Z < best.Z {
			if p.X >= float64(box.Min.X) && p.Y >= float64(box.Min.Y) {
				if p.X <= float64(box.Max.X) && p.Y <= float64(box.Max.Y) {
					best = p
				}
			}
		}

		return true
	})

	return best
}

func PrepBoundingRectForSearch() *image.Rectangle {
	return &image.Rectangle{
		Min: image.Point{1000000, 1000000},
		Max: image.Point{-1000000, -1000000},
	}
}

func BoundingRectMinMax(r *image.Rectangle, p r3.Vector) {
	x := int(p.X)
	y := int(p.Y)

	if x < r.Min.X {
		r.Min.X = x
	}
	if x > r.Max.X {
		r.Max.X = x
	}
	if y < r.Min.Y {
		r.Min.Y = y
	}
	if y > r.Max.Y {
		r.Max.Y = y
	}

}

func InBox(pt, min, max r3.Vector) bool {
	if pt.X < min.X || pt.X > max.X {
		return false
	}

	if pt.Y < min.Y || pt.Y > max.Y {
		return false
	}

	if pt.Z < min.Z || pt.Z > max.Z {
		return false
	}

	return true
}

func PCCrop(pc pointcloud.PointCloud, min, max r3.Vector) pointcloud.PointCloud {
	return PCCropWithColor(pc, min, max, nil)
}

type ColorFilter struct {
	Color    color.RGBA
	Distance float64
}

func EuclideanRGB(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// RGBA() returns uint32 in range [0, 65535], convert to [0, 255]
	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	dr := int(r1) - int(r2)
	dg := int(g1) - int(g2)
	db := int(b1) - int(b2)

	return math.Sqrt(float64(dr*dr + dg*dg + db*db))
}

func PCCropWithColor(pc pointcloud.PointCloud, min, max r3.Vector, colorFilters []ColorFilter) pointcloud.PointCloud {

	fixed := pointcloud.NewBasicEmpty()

	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		if !InBox(p, min, max) {
			return true
		}

		for _, cf := range colorFilters {
			dis := EuclideanRGB(cf.Color, d.Color())
			if dis > cf.Distance {
				return true
			}
		}

		fixed.Set(p, d)
		return true
	})

	return fixed
}

func PCToImage(pc pointcloud.PointCloud) image.Image {

	md := pc.MetaData()

	r := image.Rect(
		int(math.Floor(md.MinX)),
		int(math.Floor(md.MinY)),
		int(math.Ceil(md.MaxX)),
		int(math.Ceil(md.MaxY)),
	)

	xScale := 0
	yScale := 0

	if r.Min.X < 0 {
		xScale = -1 * r.Min.X
		r.Min.X += xScale
		r.Max.X += xScale
	}

	if r.Min.Y < 0 {
		yScale = -1 * r.Min.Y
		r.Min.Y += yScale
		r.Max.Y += yScale
	}

	if r.Max.X <= 0 {
		r.Max.X = 1
	}

	if r.Max.Y <= 0 {
		r.Max.Y = 1
	}

	img := image.NewRGBA(r)

	bestZ := map[string]float64{}

	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		x := int(p.X) + xScale
		y := int(p.Y) + yScale

		key := fmt.Sprintf("%d-%d", x, y)
		oldZ, ok := bestZ[key]
		if ok {
			if p.Z < oldZ {
				return true
			}
		}

		img.Set(x, y, d.Color())

		bestZ[key] = p.Z

		return true
	})

	return img
}

// GetApproachPoint
func GetApproachPoint(p r3.Vector, deltaLinear float64, o *spatialmath.OrientationVectorDegrees) r3.Vector {
	d := math.Pow((o.OX*o.OX)+(o.OY*o.OY)+(o.OZ*o.OZ), .5)

	xLinear := (o.OX * deltaLinear / d)
	yLinear := (o.OY * deltaLinear / d)
	zLinear := (o.OZ * deltaLinear / d)

	approachPoint := r3.Vector{
		X: p.X - xLinear,
		Y: p.Y - yLinear,
		Z: p.Z - zLinear,
	}

	return approachPoint
}

func GetMergedPointCloud(ctx context.Context, positions []toggleswitch.Switch, sleepTime time.Duration, src camera.Camera, extra map[string]interface{}, logger logging.Logger) (pointcloud.PointCloud, error) {
	logger.Infof("in GetMergedPointCloud")
	inputs := []pointcloud.PointCloud{}
	totalSize := 0

	for _, p := range positions {
		logger.Infof("about to move to position")
		err := p.SetPosition(ctx, 2, nil)
		if err != nil {
			return nil, err
		}

		time.Sleep(sleepTime)
		logger.Info("about to take point cloud image")
		pc, err := src.NextPointCloud(ctx, extra)
		if err != nil {
			return nil, err
		}

		totalSize += pc.Size()

		inputs = append(inputs, pc)
	}

	big := pointcloud.NewBasicPointCloud(totalSize)

	for _, pc := range inputs {
		err := pointcloud.ApplyOffset(pc, nil, big)
		if err != nil {
			return nil, err
		}
	}

	return big, nil
}
