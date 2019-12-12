package lia

import (
	"fmt"
	"math"
)

// Solve Domain
func (d *Domain) Solve(RuntimeSec float64) map[int]float64 {
	d.tcum = 0.
	d.dt = RuntimeSec

	prn := func() map[int]float64 {
		mout, wbal := make(map[int]float64, d.nc), 0.
		for k, n := range d.ns {
			if k > d.nc { ///////////////////////////  dic are not ordered assumes orderd map
				break
			}
			mout[k] = math.Max(n.h, n.z)
			wbal += math.Max(n.h-n.z, 0.)
		}
		fmt.Println("")
		fmt.Println(wbal)
		return mout
	}
	prn()
	// ks := []int{}
	// for k := range d.fs {
	// 	if d.bf[k] {
	// 		continue
	// 	}
	// 	ks = append(ks, k)
	// }
	// sf := make(map[int]face, len(d.fs))
	// for k, f := range d.fs {
	// 	if d.bf[k] {
	// 		continue
	// 	}
	// 	sf[k] = f
	// }
	for {
		d.setCurrentState()
		d.tcum += d.dt
		if d.tcum > RuntimeSec {
			d.dt -= d.tcum - RuntimeSec
			d.tcum = RuntimeSec
		}
		d.updateFluxes()
		// for _, k := range ks {
		// 	d.fs[k].updateFlux(&d.st[k], d.dt)
		// }
		// for k, f := range sf {
		// 	lsk := d.st[k]
		// 	f.updateFlux(&lsk, d.dt) ////////////////// GO
		// 	sf[k] = f
		// }
		d.updateHeads()
		// fmt.Print(".")

		if d.tcum == RuntimeSec {
			break
		}
	}
	return prn()
}
