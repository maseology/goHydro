package gwru

import (
	"log"
	"math"

	"github.com/maseology/goHydro/tem"
)

// New constructor. unit-volum inputs (i.e., [m/ts])
//  ksat: saturated hydraulic conductivity [m/ts]
func (t *TMQ) New(ksat map[int]float64, upcnt map[int]int, topo *tem.TEM, cw, m, Q0 float64) (map[int]float64, float64) {
	// checkInputs(ksat, topo, cw, 1., 1., m)
	t.M = m                                               // parameter [m]
	n := len(ksat)                                        // number of cells
	t.Ca = cw * cw * float64(n)                           // cw: cell width, Ca: basin (catchment) area [m²]
	strmcthresh := int(strmkm2 * 1000. * 1000. / cw / cw) // "stream cell" threshold

	// compute unit contributing areas
	uca := make(map[int]int, n)
	for i := range ksat {
		uca[i] = 1
		for _, c := range topo.UpIDs(i) {
			if _, ok := ksat[c]; ok { // to be kept within sws
				uca[i] += topo.UnitContributingArea(c)
			}
		}
	}

	g := 0.
	ti := make(map[int]float64, n)  // soil-topographic index
	t.d = make(map[int]float64, n)  // cell deficits relative to Dm
	t.Qs = make(map[int]float64, n) // saturated lateral discharge [m/ts]
	for i, k := range ksat {
		tsat := k * cw                        // lateral transmisivity when soil is saturated [m²/ts]
		tanbeta := math.Tan(topo.TEC[i].S)    // gradient
		ai := float64(uca[i]) * cw            // contributing area per unit contour [m] (assumes uniform square cells)
		ti[i] = math.Log(ai / tsat / tanbeta) // soil-topographic index
		if upcnt[i] >= strmcthresh {          // selecting only stream cells
			// t.Qs[i] = omega * tsat * tanbeta * cw / t.Ca // (Qi) saturated lateral discharge (when Dm=0) at stream cells [m/ts]
			t.Qs[i] = omega * tsat * tanbeta / cw // (Qi) saturated lateral discharge (when Dm=0) at stream cells [m/ts]
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

	// imitialize
	t.Dm = -m * (g + math.Log(Q0))
	t.steady()
	// fmt.Printf("  catchemnt area: %.3f km²; m %.3f; Dm0: %.3f; niter: %d\n", t.Ca/1000./1000., m, t.Dm, t.steady())
	return ti, g
}

func (t *TMQ) steady() (niter int) {
	niter = 0
	tl := 0.
	for {
		niter++
		qb, lsum, lcnt := 0., 0., 0.
		for i, v := range t.Qs {
			di := t.Dm + t.d[i]
			qb += v * math.Exp(-di/t.M)
		}
		qb /= float64(len(t.d))
		if math.Abs(tl-qb) < 1.e-3 || niter > 100 {
			break
		}
		for _, d := range t.d {
			di := t.Dm + d
			if di < 0 {
				lsum -= di
				lcnt++
			}
		}
		t.Dm += qb - .001 //  adding 1mm/ts recharge
		if lcnt > 0 {
			t.Dm += lsum / lcnt
		}
		if math.IsNaN(t.Dm) {
			log.Fatalf(" TMQ.steady error: Dm=NaN; m=%.3e; niter=%d\n", t.M, niter)
		}
		tl = qb
	}
	return
}
