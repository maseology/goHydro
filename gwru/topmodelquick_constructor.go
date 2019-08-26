package gwru

import (
	"log"
	"math"

	"github.com/maseology/goHydro/tem"
)

const qo = 0.25    // Qo is the discharge when basin is fully saturated (i.e., when Dm=0) [m/yr] (only needed when initializing model)
const strmkm2 = 1. // total drainage area [km²] required to deem a cell a "stream cell"

// New constructor. unit-volum inputs (i.e., [m/ts])
//  ksat: saturated hydraulic conductivity [m/ts]
func (t *TMQ) New(ksat map[int]float64, upcnt map[int]int, topo tem.TEM, cw, m, ts float64) (map[int]float64, float64) {
	checkInputs(ksat, topo, cw, 1., 1., m)
	t.M = m                                               // parameter [m]
	t.Dm = -m * math.Log(3.)                              // initialize basin-wide deficit and cell deficits [m] (assumes q0/qo~3)
	n := len(ksat)                                        // number of cells
	t.Ca = cw * cw * float64(n)                           // cw: cell width, Ca: basin (catchment) area [m²]
	strmcthresh := int(strmkm2 * 1000. * 1000. / cw / cw) // "stream cell" threshold

	g := 0.
	ti := make(map[int]float64, n)  // soil-topographic index
	t.d = make(map[int]float64, n)  // cell deficits relative to Dm
	t.Qs = make(map[int]float64, n) // saturated lateral discharge [m/ts]
	for i, k := range ksat {
		tsat := k * cw                          // lateral transmisivity when soil is saturated [m²/ts]
		tanbeta := math.Tan(topo.TEC[i].S)      // gradient
		ai := topo.UnitContributingArea(i) * cw // contributing area per unit contour [m] (assumes uniform square cells)
		ti[i] = math.Log(ai / tsat / tanbeta)   // soil-topographic index
		if upcnt[i] >= strmcthresh {
			t.Qs[i] = (cw * tsat * tanbeta / ai) // saturated lateral discharge [m/ts]
		}
		if math.IsNaN(ti[i]) {
			log.Fatalf(" TMQ.New error: topographic index is NaN. slope = %f\n", topo.TEC[i].S)
		}
		g += ti[i] // gamma
	}
	g /= float64(n) // assumes uniform square cells
	for i, v := range ti {
		t.d[i] = m * (g - v) // deficit at cell i relative to Dm [m]
	}
	t.steady(ts)
	// fmt.Printf("  catchemnt area: %.3f km²; niter: %d\n", t.Ca/1000./1000., t.steady())
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
		M:  m,
		Ca: t.Ca,
	}
}

func (t *TMQ) steady(ts float64) (niter int) {
	tl, qot := math.MaxFloat64, qo*ts/365.24/86400. // [m/yr] to [m/ts]
	niter = 0
	for {
		niter++
		tsum := 0.
		for i := range t.d {
			if t.d[i] < 0. {
				tsum -= t.d[i]
				t.d[i] = 0.
			}
		}
		bf := qot * math.Exp(-t.Dm/t.M) // [m/ts]
		if math.Abs(tl-bf) < 1.e-3 {
			break
		}
		tl = bf
	}
	return
}
