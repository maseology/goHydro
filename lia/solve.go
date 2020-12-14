package lia // VERSION 2

import (
	"fmt"
	"math"
)

const (
	dhmin = 1.e-5 // minimum of maximum global head change threshold (for steady-state simulations only)
	dmin  = 1.e-5 // minimum depth for variable timestep calculation
	g     = 9.80665
)

func (d *Domain) wbal() float64 {
	wbal := 0.
	for i := 0; i < d.GF.GD.Nact; i++ {
		wbal += math.Max(d.ns[i].h-d.ns[i].z, 0.)
	}
	return wbal
}

func (d *Domain) mh() map[int]float64 {
	mout := make(map[int]float64, d.GF.GD.Nact)
	for i, c := range d.GF.GD.Sactives {
		n := d.ns[i]
		mout[c] = math.Max(n.h, n.z)
	}
	return mout
}

func (d *Domain) prn() (map[int]float64, float64) {
	mout, wbal := make(map[int]float64, d.GF.GD.Nact), 0.
	for i, c := range d.GF.GD.Sactives {
		n := d.ns[i]
		mout[c] = math.Max(n.h, n.z)
		wbal += math.Max(n.h-n.z, 0.)
	}
	return mout, wbal
}

// Solve domain to a specified time, constant BC
func (d *Domain) Solve(RuntimeSec float64) map[int]float64 {
	d.tcum = 0.
	d.dt = RuntimeSec
	// wbal0 := d.wbal()
	for {
		d.setDT()
		if d.tcum > RuntimeSec {
			d.dt -= d.tcum - RuntimeSec
			d.tcum = RuntimeSec
		}

		d.updateFluxes()
		d.updateHeads()
		// fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
		fmt.Print(".")

		if d.tcum == RuntimeSec {
			break
		}
	}
	mh, _ := d.prn()
	// bias := (wbal1 - wbal0) / wbal0
	// if math.Abs(bias) > 1.e-5 {
	// 	fmt.Println(" ***** mass balance issue, consider lowering alpha")
	// }
	// fmt.Printf("mass balance %f%%\n", bias*100.)
	return mh
}

// SolveSteadyState domain to a steady state
func (d *Domain) SolveSteadyState() map[int]float64 {
	// wbal0 := d.wbal()
	for {
		d.setDT()
		d.updateFluxes()
		dhx := d.updateHeads()
		// fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
		fmt.Print(".")
		if math.Abs(dhx) < dhmin {
			break
		}
	}
	fmt.Print("\n")
	mh, _ := d.prn()
	// bias := (wbal1 - wbal0) / wbal0
	// if math.Abs(bias) > 1.e-5 {
	// 	fmt.Println(" ***** mass balance issue, consider lowering alpha")
	// }
	// fmt.Printf("mass balance %f%%\n", bias*100.)
	return mh
}

// SolveTransient solves the solution with a transient head generation function
// currently only returns the final state
// used mainly when comparing to analytical solutions
func (d *Domain) SolveTransient(RuntimeSec, ReportSec float64, gen func(float64) map[int]float64) []map[int]float64 {
	d.tcum = 0.
	d.dt = RuntimeSec
	rprt, irprt := 0., 1
	hcol, wbal0 := newColl(int(RuntimeSec), int(ReportSec)), 0.
	hcol[0], wbal0 = d.prn()
	for {
		d.SetHeads(gen(d.tcum))
		// for f, h := range gen(d.tcum) {
		// 	d.gns[f].h = h
		// }
		d.setDT()
		if d.tcum > RuntimeSec {
			d.dt -= d.tcum - RuntimeSec
			d.tcum = RuntimeSec
		}
		if d.tcum > rprt {
			d.dt -= d.tcum - rprt
			d.tcum = rprt
		}

		d.updateFluxes()
		dhx := d.updateHeads()
		fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
		// fmt.Print(".")

		if d.tcum == rprt {
			hcol[irprt] = d.mh()
			rprt += ReportSec
			irprt++
		}
		if d.tcum == RuntimeSec {
			break
		}
	}
	// fmt.Printf("\n")
	bias := (d.wbal() - wbal0) / wbal0
	if math.Abs(bias) > 1.e-5 {
		fmt.Println(" ***** mass balance issue, consider lowering alpha")
	}
	fmt.Printf("mass balance %f%%\n", bias*100.)
	return hcol
}

func newColl(RuntimeSec, ReportSec int) []map[int]float64 {
	if RuntimeSec%ReportSec != 0 {
		print("newColl: uneven timesteps")
		return nil
	}
	return make([]map[int]float64, RuntimeSec/ReportSec+2)
}

// // SolveSteadyState domain to a steady state
// func (d *Domain) SolveSteadyState() map[int]float64 {
// 	_, wbal0 := d.prn()
// 	d.tcum = 0.
// 	iter := 0
// 	for {
// 		d.setDT()

// 		d.updateFluxes()
// 		dhx := d.updateHeads()
// 		fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
// 		if math.Abs(dhx) < dhmin {
// 			break
// 		}
// 		iter++
// 	}
// 	mh, wbal1 := d.prn()
// 	bias := (wbal1 - wbal0) / wbal0
// 	if math.Abs(bias) > 1.e-5 {
// 		fmt.Println(" ***** mass balance issue, consider lowering alpha")
// 	}
// 	fmt.Printf("mass balance %f%%\n", bias*100.)
// 	return mh
// }

// // Solve domain to a specified time, constant BC
// func (d *Domain) Solve(RuntimeSec float64) map[int]float64 {
// 	_, wbal0 := d.prn()
// 	d.tcum = 0.
// 	d.dt = RuntimeSec
// 	iter := 0
// 	for {
// 		d.setDT()
// 		if d.tcum > RuntimeSec {
// 			d.dt -= d.tcum - RuntimeSec
// 			d.tcum = RuntimeSec
// 		}

// 		d.updateFluxes()
// 		dhx := d.updateHeads()
// 		fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
// 		// fmt.Print(".")

// 		if d.tcum == RuntimeSec {
// 			break
// 		}
// 		iter++
// 	}
// 	mh, wbal1 := d.prn()
// 	bias := (wbal1 - wbal0) / wbal0
// 	if math.Abs(bias) > 1.e-5 {
// 		fmt.Println(" ***** mass balance issue, consider lowering alpha")
// 	}
// 	fmt.Printf("mass balance %f%%\n", bias*100.)
// 	return mh
// }

// // SolveTransient solves the solution with a transient head generation function
// func (d *Domain) SolveTransient(RuntimeSec float64, gen func(float64) map[int]float64) map[int]float64 {
// 	_, wbal0 := d.prn()
// 	d.tcum = 0.
// 	d.dt = RuntimeSec
// 	iter := 0
// 	for {
//      // try: d.SetHeads(gen(d.tcum))
// 		for f, h := range gen(d.tcum) {
// 			d.gns[f].h = h
// 		}

// 		d.setDT()
// 		if d.tcum > RuntimeSec {
// 			d.dt -= d.tcum - RuntimeSec
// 			d.tcum = RuntimeSec
// 		}

// 		d.updateFluxes()
// 		dhx := d.updateHeads()
// 		fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
// 		// fmt.Print(".")

// 		if d.tcum == RuntimeSec {
// 			break
// 		}
// 		iter++
// 	}
// 	mh, wbal1 := d.prn()
// 	bias := (wbal1 - wbal0) / wbal0
// 	if math.Abs(bias) > 1.e-5 {
// 		fmt.Println(" ***** mass balance issue, consider lowering alpha")
// 	}
// 	fmt.Printf("mass balance %f%%\n", bias*100.)
// 	return mh
// }
