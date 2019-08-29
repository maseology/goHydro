package gwru

import (
	"fmt"
	"log"
	"math"

	"github.com/maseology/goHydro/tem"
)

// New constructor. unit-volum inputs (i.e., [m/ts])
//  ksat: saturated hydraulic conductivity [m/ts]
func (t *TMQ) New(ksat map[int]float64, upcnt map[int]int, topo tem.TEM, cw, m, ts float64) (map[int]float64, float64) {
	checkInputs(ksat, topo, cw, 1., 1., m)
	// t.Qo = qo * ts / 365.24 / 86400. // [m/yr] to [m/ts]
	t.M = m // parameter [m]
	// t.Dm = -m * math.Log(3.)                              // initialize basin-wide deficit and cell deficits [m] (assumes q0/qo~3)
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
		if upcnt[i] >= strmcthresh {            // selecting only stream cells
			t.Qs[i] = omega * tsat * tanbeta // (qi) saturated lateral discharge (when Dm=0) at stream cells [m²/ts]
		}
		if math.IsNaN(ti[i]) {
			log.Fatalf(" TMQ.New error: topographic index is NaN. slope = %f\n", topo.TEC[i].S)
		}
		g += ti[i] // gamma
	}

	ccc := 0
	asum := 0.
	for c0 := range t.Qs {
		cnt, asumi := 0, 0.
		for c1 := range topo.UpIDs(c0) {
			if _, ok := t.Qs[c1]; !ok {
				ccc += topo.UpCnt(c1)
				asumi += topo.UnitContributingArea(c1)
				cnt++
			}
		}
		// fmt.Printf(" %d", cnt)
		asum += asumi //* cw // lateral contributing area (to stream cells) per unit contour [m] (assumes uniform square cells)
	}
	if asum != t.Ca {
		fmt.Printf("%d (%d) %f %f\n", n, len(t.Qs), t.Ca, asum)
		log.Fatalf(" TMQ.New error: catchment areal calculation error: Ca: %.2e  ai*w: %.2e\n", t.Ca, asum*cw)
	}

	g /= float64(n) // assumes uniform square cells
	cd := 0.
	for i, v := range ti {
		t.d[i] = m * (g - v) // deficit at cell i relative to Dm [m]
		if t.d[i] < cd {
			cd = t.d[i] // initialize without ponding
		}
	}
	// t.Dm = 1.75
	fmt.Printf("  catchemnt area: %.3f km²; Dm0: %.3f; niter: %d\n", t.Ca/1000./1000., t.Dm, t.steady(ts))
	return ti, g
}

func (t *TMQ) steady(ts float64) (niter int) {
	tl, g := math.MaxFloat64, gyr*ts/365.24/86400. // [m/yr] to [m/ts]
	t.Dm = -t.M * math.Log(3.)                     // initialize basin-wide deficit and cell deficits [m] (assumes q0/qo~3)
	niter = 0
	for {
		niter++
		qb := 2. * g * math.Exp(-t.Dm/t.M) // [m/ts] (assumes max monthly baseflow rate (Qo) is twice annual average recharge)
		if math.Abs(tl-qb) < 1.e-3 {
			break
		}
		t.Dm += qb - g // remove baseflow discharge and add steady recharge
		tl = qb
	}
	return
}
