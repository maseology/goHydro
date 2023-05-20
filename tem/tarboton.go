package tem

import (
	"math"
)

var (
	re1   = [8]int{0, -1, -1, 0, 0, 1, 1, 0}
	ce1   = [8]int{1, 0, 0, -1, -1, 0, 0, 1}
	re2   = [8]int{-1, -1, -1, -1, 1, 1, 1, 1}
	ce2   = [8]int{1, 1, -1, -1, -1, -1, 1, 1}
	ac    = [8]float64{0, 1, 1, 2, 2, 3, 3, 4}
	af    = [8]float64{1, -1, 1, -1, 1, -1, 1, -1}
	atan1 = math.Atan(1.)
)

func gridSlopeAspectTarboton(bufz map[int]float64, cid0, ncol int, cw float64) (float64, float64) {
	//  ref: Tarboton D.G., 1997. A new method for the determination of flow directions and upslope areas in grid digital elevation models. Water Resources Research 33(2). p.309-319.
	//  triangular facets, ordered by steepest (assumes uniform cells)
	//  facets (slightly modified from Tarboton, 1997):
	//         \2|1/
	//         3\|/0
	//         --+--
	//         4/|\7
	//         /5|6\

	e0 := func() float64 {
		if z, ok := bufz[cid0]; ok {
			return z
		}
		panic("gridSlopeAspectTarboton err0")
	}()

	var e1, e2 float64
	var ok bool
	hcw := math.Sqrt(2 * cw * cw)
	sx, rx, kx := 0., -9999., -1
	for k := 0; k < 8; k++ {
		if e1, ok = bufz[cid0+re1[k]*ncol+ce1[k]]; !ok {
			continue
		}
		if e2, ok = bufz[cid0+re2[k]*ncol+ce2[k]]; !ok {
			continue
		}
		if e1 > e0 && e2 > e0 {
			continue
		}
		s1 := (e0 - e1) / cw
		s2 := (e1 - e2) / cw
		r := math.Atan(s2 / s1)
		s := math.Sqrt(s1*s1 + s2*s2)
		if r < 0 {
			r = 0
			s = s1
		} else if r > atan1 {
			r = atan1
			s = (e0 - e2) / hcw
		}
		if s > sx {
			sx = s
			rx = r
			kx = k
		}
	}
	if kx < 0 {
		return 0., -9999.
	}
	rg := af[kx]*rx + ac[kx]*math.Pi/2.
	if rg > math.Pi {
		rg -= 2 * math.Pi // [-pi,pi]
	}
	return sx, rg
}
