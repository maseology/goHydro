package gwru

import (
	"log"
	"math"

	"github.com/maseology/goHydro/tem"
)

// New constructor
func (t *TMQ) New(ksat map[int]float64, topo tem.TEM, cw, q0, qo, m float64) (map[int]float64, float64) {
	// ksat: saturated hydraulic conductivity [m/ts]
	// q0: initial catchment flow rate [m³/ts]
	checkInputs(ksat, topo, cw, q0, qo, m)
	t.M = m                     // parameter [m]
	t.Qo = qo                   // qo: baseflow when basin is fully saturated [m³/ts]
	t.Dm = -m * math.Log(q0/qo) // initialize basin-wide deficit and cell deficits [m]
	n := len(ksat)              // number of cells
	t.Ca = cw * cw * float64(n) // cw: cell width, ca: basin area [m²]

	g := 0.
	ti := make(map[int]float64, n)
	t.t = make(map[int]float64, n)
	for i, k := range ksat {
		t0 := k * cw                                        // lateral transmisivity when soil is saturated [m²/ts]
		ai := topo.UnitContributingArea(i) * cw             // contributing area per unit contour [m] (assumes uniform square cells)
		ti[i] = math.Log(ai / t0 / math.Tan(topo.TEC[i].S)) // soil-topographic index
		if math.IsNaN(ti[i]) {
			log.Fatalf(" TMQ.New error: topographic index is NaN. slope = %f\n", topo.TEC[i].S)
		}
		g += ti[i] // gamma
	}
	g /= float64(n) // assumes uniform square cells
	for i, v := range ti {
		t.t[i] = m * (g - v)
	}
	return ti, g
}

// Clone creates a deep copy of TMQ, while changing recession coefficient m
func (t *TMQ) Clone(m float64) TMQ {
	tnew := make(map[int]float64, len(t.t))
	for i, v := range t.t {
		tnew[i] = m * v / t.M
	}
	return TMQ{
		t:  tnew,
		Dm: 0.,
		Qo: t.Qo,
		M:  m,
		Ca: t.Ca,
	}
}
