package tem

import "github.com/maseology/goHydro/grid"

// NewTEM loads TEM
func NewFromReal(r grid.Real) (*TEM, error) {
	var t TEM

	bufs := r.GD.Buffers(false, true)

	t.TEC = buildTECs(r, bufs)
	// t.TEC = make(map[int]TEC, r.GD.Nact)
	// for c, z := range r.A {
	// 	bufz := make(map[int]float64, 9)
	// 	bufz[c] = z
	// 	for _, cc := range bufs[c] {
	// 		if cc >= 0 {
	// 			if !math.IsInf(r.A[cc], 0) {
	// 				bufz[cc] = r.A[cc]
	// 			}
	// 		}
	// 	}

	// 	g, a := gridSlopeAspectTarboton(bufz, c, r.GD.Ncol, r.GD.Cwidth)
	// 	t.TEC[c] = TEC{Z: z, G: g, A: a}
	// }

	ds := t.buildDsFromNeighbours(bufs)
	// t.checkVals()
	t.buildUpslopes(ds)

	return &t, nil
}

func BuildTECs(r grid.Real) map[int]TEC {
	bufs := r.GD.Buffers(false, true)
	return buildTECs(r, bufs)
}
