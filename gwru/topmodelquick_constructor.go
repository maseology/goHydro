package gwru

import (
	"math"

	"github.com/maseology/goHydro/tem"
)

// New constructor
func (t *TMQ) New(ksat map[int]float64, topo tem.TEM, cw, q0, qo, m float64) {
	// ksat: saturated hydraulic conductivity [m/ts]
	// q0: initial catchment flow rate [m³/ts]
	checkInputs(ksat, topo, cw, q0, qo, m)
	t.m = m                     // parameter [m]
	t.qo = qo                   // qo: baseflow when basin is fully saturated [m3/ts]
	n := topo.NumCells()        // number of cells
	t.ca = cw * cw * float64(n) // cw: cell width, ca: basin area [m2]

	g := 0.
	ti := make(map[int]float64, n)
	t.t = make(map[int]float64, n)
	for i, p := range topo.TEC {
		t0 := ksat[i] * cw                        // lateral transmisivity when soil is saturated [m²/ts]
		ai := topo.UnitContributingArea(i) / cw   // contributing area per unit contour [m]
		ti[i] = math.Log(ai / t0 / math.Tan(p.S)) // soil-topographic index
		g += ti[i]                                // gamma
	}
	g /= float64(n)
	for i, v := range ti {
		t.t[i] = m * (g - v)
	}
	t.Dm = -m * math.Log(q0/qo) // initialize basin-wide deficit and cell deficits [m]
}
