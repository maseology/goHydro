package tem

import (
	"math"

	"github.com/maseology/goHydro/grid"
)

func buildTECs(r grid.Real, bufs map[int][]int) map[int]TEC {
	tec := make(map[int]TEC, r.GD.Nact)
	for c, z := range r.A {
		bufz := make(map[int]float64, 9)
		bufz[c] = z
		for _, cc := range bufs[c] {
			if cc >= 0 {
				if !math.IsInf(r.A[cc], 0) {
					bufz[cc] = r.A[cc]
				}
			}
		}
		g, a := gridSlopeAspectTarboton(bufz, c, r.GD.Ncol, r.GD.Cwidth)
		tec[c] = TEC{Z: z, G: g, A: a}
	}
	return tec
}
