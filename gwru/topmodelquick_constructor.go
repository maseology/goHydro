package gwru

import (
	"log"
	"math"

	"github.com/maseology/goHydro/tem"
)

// New constructor. unit-volum inputs (i.e., [m/ts])
//  ksat: saturated hydraulic conductivity [m/ts]
//  q0: initial catchment flow rate [m/ts]
func (t *TMQ) New(ksat map[int]float64, topo tem.TEM, cw, qo, m float64) (map[int]float64, float64) {
	checkInputs(ksat, topo, cw, 1., qo, m)
	t.M = m                     // parameter [m]
	t.Qo = qo                   // qo: baseflow when basin is fully saturated [m/ts]
	t.Dm = -m * math.Log(3.)    // initialize basin-wide deficit and cell deficits [m] (assumes q0/qo~3)
	n := len(ksat)              // number of cells
	t.Ca = cw * cw * float64(n) // cw: cell width, Ca: basin (catchment) area [m²]

	g := 0.
	ti := make(map[int]float64, n) // soil-topographic index
	t.d = make(map[int]float64, n) // cell deficits
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
		t.d[i] = m * (g - v) // deficit at cell i [m]
	}
	t.steady()
	return ti, g
}

// Clone creates a deep copy of TMQ, while changing recession coefficient m
func (t *TMQ) Clone(m float64) TMQ {
	dnew := make(map[int]float64, len(t.d))
	for i, v := range t.d {
		dnew[i] = m * v / t.M
	}
	return TMQ{
		d:  dnew,
		Dm: 0.,
		Qo: t.Qo,
		M:  m,
		Ca: t.Ca,
	}
}

func (t *TMQ) steady() {
	tl := math.MaxFloat64
	for {
		tsum := 0.
		for i := range t.d {
			if t.d[i] < 0. {
				tsum -= t.d[i]
				t.d[i] = 0.
			}
		}
		bf := t.Update(t.Qo) // [m/ts]
		if math.Abs(tl-bf) < 1.e-3 {
			break
		}
		tl = bf
	}
}
