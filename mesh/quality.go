package mesh

import (
	"math"

	"github.com/maseology/mmaths"
)

type MeshQuality struct {
	Area, MaxAngle, Skewness, AspectRatio float64
	DelaunayFail, IsRight, IsObtuse       bool
}

func (sl *Slice) Quality() []MeshQuality {
	dist2D := func(b, e []float64) float64 {
		return math.Sqrt(math.Pow(b[0]-e[0], 2.) + math.Pow(b[1]-e[1], 2.))
	}
	aspectRatio := func(v [][]float64) float64 {
		l0 := dist2D(v[0], v[1])
		l1 := dist2D(v[1], v[2])
		l2 := dist2D(v[2], v[0])
		return math.Max(l0, math.Max(l1, l2)) / math.Min(l0, math.Min(l1, l2))
	}

	circums := make([]*mmaths.Circle, len(sl.Elements))
	o := make([]MeshQuality, len(sl.Elements))
	te := math.Pi / 3. // the angle for equi-angular face or cell i.e. 60° for a triangle and 90° for a square.
	for eid, el := range sl.Elements {
		if len(el) != 3 {
			panic("mesh.Quality: assumes triangular elements")
		}
		xys := make([][]float64, 3)
		for i, nd := range el {
			xys[i] = sl.Nodes[nd]
		}
		tri := sl.ElementToTriangle(eid)
		circums[eid] = tri.Circumcircle()
		an, ax := tri.MinMaxInteriorAngle()
		o[eid] = MeshQuality{
			Area:         tri.Area(),
			MaxAngle:     ax,
			Skewness:     math.Max((ax-te)/(math.Pi-te), (te-an)/te), // the skewness of a grid is an apt indicator of the mesh quality and suitability. Large skewness compromises the accuracy of the interpolated regions,
			AspectRatio:  aspectRatio(xys),                           // the ratio of longest to the shortest side in a cell. Ideally it should be equal to 1 to ensure best results
			DelaunayFail: false,                                      // initialize
			IsRight:      ax == math.Pi/2.,
			IsObtuse:     ax > math.Pi/2.,
		}
	}

	// check Delaunay criterion
	for _, edg := range sl.GetEdges() {
		if edg.E1 < 0 { // boundary element
			continue
		}
		nchk := func() int { // find node ID other than either pair from adjacent element
			for _, nid := range sl.Elements[edg.E1] {
				if nid == edg.N0 || nid == edg.N1 {
					continue
				}
				return nid
			}
			return -1
		}()
		if circums[edg.E0].Contains(sl.Nodes[nchk]) {
			o[edg.E0].DelaunayFail = true
			o[edg.E1].DelaunayFail = true
		}
	}
	return o
}
