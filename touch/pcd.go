package touch

import (
	"image"

	"github.com/golang/geo/r3"

	"go.viam.com/rdk/pointcloud"
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
