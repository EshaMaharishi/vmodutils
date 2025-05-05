package touch

func PCDFindHighestInRegion(p pointcloud.PointCloud, box image.Rectangle) r3.Vector {

	best := r3.Vector{Z: -100000}

	pc.Iterate(0, 0, func(p r3.Vector, d pointcloud.Data) bool {
		if p.Z > best.Z {
			if p.X >= box.Min.X && p.Y >= box.Min.Y {
				if p.X <= box.Max.X && p.Y >= box.Max.Y {
					best = p
				}
			}
		}

		return true
	})

	return best
}
