package gwru

import (
	"fmt"
	"log"
	"math"

	"github.com/maseology/goHydro/tem"
)

// New constructor
func (t *TOPMODEL) New(ksat map[int]float64, topo tem.TEM, cw, q0, qo, m float64) {
	// q0: initial catchment flow rate [m³/s]
	checkInputs(ksat, topo, cw, q0, qo, m)
	t.m = m                       // parameter m
	t.qo = qo                     // qo: baseflow when basin is fully saturated [m3/s]
	n := float64(topo.NumCells()) // number of cells
	t.ca = cw * cw * n            // cw: cell width, ca: basin area [m2]

	t.g = 0.                                // gamma
	t.ti = make(map[int]float64, len(ksat)) // soil-topographic index
	t.Di = make(map[int]float64, len(ksat)) // depth to watertable
	for i, k := range ksat {
		t0 := k * cw                            // lateral transmisivity when soil is saturated [m²/s]
		ai := topo.UnitContributingArea(i) / cw // contributing area per unit contour [m]
		t.ti[i] = math.Log(ai / t0 / math.Tan(topo.TECs[i].S))
		t.g += t.ti[i]
	}
	t.g /= n
	dm := -t.m * math.Log(q0/qo) // initialize basin-wide deficit and cell deficits
	t.updateDeficits(dm)
}

func checkInputs(ksat map[int]float64, topo tem.TEM, cw, q0, qo, m float64) {
	for i, k := range ksat {
		if v, ok := topo.TECs[i]; ok {
			if k <= 0. {
				log.Panicf("TOPMODEL error: cell %d has an assigned ksat = %v", i, k)
			}
			if v.S <= 0. {
				fmt.Printf("TOPMODEL warning: slope at cell %d was found to be %v, reset to 0.0001.", i, v.S)
				v.S = 0.0001
			}
		} else {
			log.Panicf("TOPMODEL error: TEC map does not contain value for cell %d", i)
		}
	}
	if m <= 0. {
		log.Panic("TOPMODEL error: parameter m must be >0.")
	}
	if qo <= 0. {
		log.Panic("TOPMODEL error: qo must be >0.")
	}
	if q0 <= 0. {
		println("TOPMODEL warning: q0 must be >0, reset to 0.001.")
		q0 = 0.001
	}
	if cw <= 0. {
		log.Panic("TOPMODEL error: cell width must be >0.")
	}
}
