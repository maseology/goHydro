package gwru

import (
	"fmt"
	"log"
	"math"

	"github.com/maseology/goHydro/tem"
)

// New constructor
func (t *TOPMODEL) New(ksat map[int]float64, topo *tem.TEM, cw, q0, qo, m float64) {
	// ksat: saturated hydraulic conductivity [m/ts]
	// q0: initial catchment flow rate [m³/ts]
	checkInputs(ksat, topo, cw, q0, qo, m)
	t.m = m                     // parameter [m]
	t.qo = qo                   // qo: baseflow when basin is fully saturated [m3/ts]
	n := topo.NumCells()        // number of cells
	t.ca = cw * cw * float64(n) // cw: cell width, ca: basin area [m2]

	t.g = 0.
	t.ti = make(map[int]float64, n)
	t.Di = make(map[int]float64, n) // soil moisture deficit (~depth to watertable * porosity)
	for i, p := range topo.TEC {
		t0 := ksat[i] * cw                               // lateral transmisivity when soil is saturated [m²/ts]
		ai := float64(topo.UnitContributingArea(i)) / cw // contributing area per unit contour [m]
		t.ti[i] = math.Log(ai / t0 / math.Tan(p.G))      // soil-topographic index
		t.g += t.ti[i]                                   // gamma
	}
	t.g /= float64(n)
	t.Dm = -t.m * math.Log(q0/qo) // initialize basin-wide deficit and cell deficits [m]
	t.updateDeficits()
}

func checkInputs(ksat map[int]float64, topo *tem.TEM, cw, q0, qo, m float64) {
	for i, k := range ksat {
		if k <= 0. {
			log.Panicf(" TOPMODEL.checkInputs error: cell %d has an assigned ksat = %v\n", i, k)
		}
		if p, ok := topo.TEC[i]; ok {
			if p.G <= 0. {
				fmt.Printf(" TOPMODEL.checkInputs warning: slope at cell %d was found to be %v, reset to 0.0001.", i, p.G)
				t := topo.TEC[i]
				t.G = 0.0001
				t.A = 0.
				topo.TEC[i] = t
			}
		} else {
			log.Panicf(" TOPMODEL.checkInputs error: no topographic info available for cell %d", i)
		}
	}

	// for i, p := range topo.TEC {
	// 	if k, ok := ksat[i]; ok {
	// 		if k <= 0. {
	// 			log.Panicf(" TOPMODEL error: cell %d has an assigned ksat = %v", i, k)
	// 		}
	// 		if p.S <= 0. {
	// 			fmt.Printf(" TOPMODEL warning: slope at cell %d was found to be %v, reset to 0.0001.", i, p.S)
	// 			t := topo.TEC[i]
	// 			t.S = 0.0001
	// 			topo.TEC[i] = t
	// 		}
	// 	} else {
	// 		log.Panicf(" TOPMODEL error: ksat map does not contain value for cell %d", i)
	// 	}
	// }
	if m <= 0. {
		log.Panic(" TOPMODEL error: parameter m must be >0.")
	}
	if qo <= 0. {
		log.Panic(" TOPMODEL error: qo must be >0.")
	}
	if q0 <= 0. {
		println(" TOPMODEL warning: q0 must be >0, reset to 0.001.")
		q0 = 0.001
	}
	if cw <= 0. {
		log.Panic(" TOPMODEL error: cell width must be >0.")
	}
}
