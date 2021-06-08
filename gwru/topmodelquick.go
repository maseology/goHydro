package gwru

import (
	"math"
)

const (
	omega = 1. // sinuosity
)

// TMQ is an optimized, distributed variation of the TOPMODEL struct
type TMQ struct {
	D, Qs map[int]float64 // cell deficit relative to Dm; saturated lateral discharge (=omega To tan(beta)/w)
	M     float64         // parameter m, gamma
}

// // Copy TMQ
// func (t *TMQ) Copy() TMQ {
// 	return TMQ{
// 		D:  mmio.CopyMapif(t.D),
// 		Qs: mmio.CopyMapif(t.Qs),
// 	}
// }

// New constructor. unit-volum inputs (i.e., [m/ts])
//  ksat: saturated hydraulic conductivity [m/ts]
//  uca: unit contributing areas to every cell
func (t *TMQ) New(ksat, grad map[int]float64, uca map[int]int, strm []int, cw, m float64) (map[int]float64, float64) {
	t.M = m
	n := len(ksat)                          // number of cells
	t.Qs = make(map[int]float64, len(strm)) // saturated lateral discharge [m/ts]
	for _, c := range strm {
		if _, ok := ksat[c]; ok {
			t.Qs[c] = 0.
		}
	}

	g := 0.
	ti := make(map[int]float64, n) // soil-topographic index
	t.D = make(map[int]float64, n) // cell deficits relative to Dm
	for i, k := range ksat {
		tsat := k * cw                        // lateral transmissivity when soil is saturated [m²/ts]
		tanbeta := math.Tan(grad[i])          // tangent slope gradient
		ai := float64(uca[i]) * cw            // contributing area per unit contour [m] (assumes uniform square cells)
		ti[i] = math.Log(ai / tsat / tanbeta) // soil-topographic index
		// if math.IsNaN(ti[i]) || math.IsInf(ti[i], 0) {
		// 	log.Fatalf(" TMQ.New error: topographic index is either NaN or Inf. slope = %f\n", topo.TEC[i].G)
		// }
		if _, ok := t.Qs[i]; ok {
			t.Qs[i] = omega * tsat * tanbeta / cw // (Qi) saturated lateral discharge (when Dm=0) at stream cells [m/ts]
		}
		g += ti[i] // gamma
	}
	g /= float64(n) // assumes uniform square cells

	for i, v := range ti {
		t.D[i] = m * (g - v) // deficit at cell i relative to Dm [m]
	}

	return ti, g
}

// RelTi returns the topographical index relative to gamma (g-Ti)
func (t *TMQ) RelTi() map[int]float64 {
	out := make(map[int]float64, len(t.D))
	for k, v := range t.D {
		out[k] = v / t.M
	}
	return out
}
