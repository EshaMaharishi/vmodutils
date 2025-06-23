package smtools

import (
	"fmt"
	"os"

	"github.com/golang/geo/r3"
	"neilpa.me/go-stl"

	"go.viam.com/rdk/spatialmath"
)

func scale(v r3.Vector, amount float64) r3.Vector {
	return r3.Vector{
		X: v.X * amount,
		Y: v.Y * amount,
		Z: v.Z * amount,
	}
}

func STLFileToGeometry(fn string) (spatialmath.Geometry, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err

	}
	defer f.Close()

	mesh, err := stl.Decode(f)
	if err != nil {
		return nil, err
	}

	minPoint := r3.Vector{100000, 10000, 1000}
	maxPoint := r3.Vector{-10000, -10000, -10000}

	triangles := []*spatialmath.Triangle{}

	for _, face := range mesh.Faces {
		if face.AttributeByteCount > 0 {
			return nil, fmt.Errorf("expect face.AttributeByteCount to be 0, but is %d", face.AttributeByteCount)
		}

		vs := []r3.Vector{}

		for _, v := range face.Verts {
			vs = append(vs, r3.Vector{float64(v[0]), float64(v[1]), float64(v[2])})

			minPoint.X = min(minPoint.X, float64(v[0]))
			minPoint.Y = min(minPoint.Y, float64(v[1]))
			minPoint.Z = min(minPoint.Z, float64(v[2]))

			maxPoint.X = max(maxPoint.X, float64(v[0]))
			maxPoint.Y = max(maxPoint.Y, float64(v[1]))
			maxPoint.Z = max(maxPoint.Z, float64(v[2]))
		}

		triangles = append(triangles, spatialmath.NewTriangle(vs[0], vs[1], vs[2]))
	}

	minPoint = scale(minPoint, 1000)
	maxPoint = scale(maxPoint, 1000)

	dims := r3.Vector{
		X: maxPoint.X - minPoint.X,
		Y: maxPoint.Y - minPoint.Y,
		Z: maxPoint.Z - minPoint.Z,
	}

	fmt.Printf("min: %v\nmax: %v\ndims: %v\n", minPoint, maxPoint, dims)

	b, err := spatialmath.NewBox(spatialmath.NewZeroPose(), dims, "")
	if err != nil {
		return nil, fmt.Errorf("couldn't create box %w", err)
	}

	m := spatialmath.NewMesh(spatialmath.NewZeroPose(), triangles, "")

	isEncompassed, err := m.EncompassedBy(b)
	if err != nil {
		return nil, fmt.Errorf("could run EncompassedBy: %w", err)
	}
	if !isEncompassed {
		return nil, fmt.Errorf("mesh not encompassed")
	}

	return b, nil
}
