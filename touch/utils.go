package touch

import (
	"fmt"
	"image"
	"math"

	"github.com/golang/geo/r3"

	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/spatialmath"
)

func PCDFindHighestInRegion(pc pointcloud.PointCloud, box image.Rectangle) r3.Vector {

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

func PCDCrop(pc pointcloud.PointCloud, min, max r3.Vector) pointcloud.PointCloud {
	fixed := pointcloud.NewBasicEmpty()

	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		if InBox(p, min, max) {
			fixed.Set(p, d)
		}
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

func GetApproachPoint(md pointcloud.MetaData, c r3.Vector, deltaLinear float64, o *spatialmath.OrientationVectorDegrees) r3.Vector {

	d := math.Pow((o.OX*o.OX)+(o.OY*o.OY), .5)

	xLinear := (o.OX * deltaLinear / d)
	yLinear := (o.OY * deltaLinear / d)

	approachPoint := r3.Vector{
		X: c.X - xLinear,
		Y: c.Y - yLinear,
		Z: c.Z,
	}

	return approachPoint
}
