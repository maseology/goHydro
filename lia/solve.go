package lia// VERSION 2

import (
	"fmt"
	"math"
)

func (d *Domain) prn() (map[int]float64, float64) {
	mout, wbal := make(map[int]float64, d.GF.GD.Na), 0.
	for i, c := range d.GF.GD.Sactives {
		n := d.ns[i]
		mout[c] = math.Max(n.h, n.z)
		wbal += math.Max(n.h-n.z, 0.)
	}
	return mout, wbal
}

// SolveSteadyState domain to a steady state
func (d *Domain) SolveSteadyState() map[int]float64 {
	_, wbal0 := d.prn()
	d.tcum = 0.
	iter := 0
	for {
		d.setDT()

		d.updateFluxes()
		dhx := d.updateHeads()
		fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
		if math.Abs(dhx) < dhmin {
			break
		}
		iter++
	}
	mh, wbal1 := d.prn()
	bias := (wbal1 - wbal0) / wbal0
	if math.Abs(bias) > 1.e-5 {
		fmt.Println(" ***** mass balance issue, consider lowering alpha")
	}
	fmt.Printf("mass balance %f%%\n", bias*100.)
	return mh
}

// Solve domain to a specified time, constant BC
func (d *Domain) Solve(RuntimeSec float64) map[int]float64 {
	_, wbal0 := d.prn()
	d.tcum = 0.
	d.dt = RuntimeSec
	iter := 0
	for {
		d.setDT()
		if d.tcum > RuntimeSec {
			d.dt -= d.tcum - RuntimeSec
			d.tcum = RuntimeSec
		}

		d.updateFluxes()
		dhx := d.updateHeads()
		fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
		// fmt.Print(".")

		if d.tcum == RuntimeSec {
			break
		}
		iter++
	}
	mh, wbal1 := d.prn()
	bias := (wbal1 - wbal0) / wbal0
	if math.Abs(bias) > 1.e-5 {
		fmt.Println(" ***** mass balance issue, consider lowering alpha")
	}
	fmt.Printf("mass balance %f%%\n", bias*100.)
	return mh
}

// SolveTransient solves the solution with a transient head generation function
func (d *Domain) SolveTransient(RuntimeSec float64, gen func(float64) map[int]float64) map[int]float64 {
	_, wbal0 := d.prn()
	d.tcum = 0.
	d.dt = RuntimeSec
	iter := 0
	for {
		for f, h := range gen(d.tcum) {
			d.gns[f].h = h
		}

		d.setDT()
		if d.tcum > RuntimeSec {
			d.dt -= d.tcum - RuntimeSec
			d.tcum = RuntimeSec
		}

		d.updateFluxes()
		dhx := d.updateHeads()
		fmt.Printf("%10.5f %10.5f %10.5f\n", d.tcum, dhx, d.dt)
		// fmt.Print(".")

		if d.tcum == RuntimeSec {
			break
		}
		iter++
	}
	mh, wbal1 := d.prn()
	bias := (wbal1 - wbal0) / wbal0
	if math.Abs(bias) > 1.e-5 {
		fmt.Println(" ***** mass balance issue, consider lowering alpha")
	}
	fmt.Printf("mass balance %f%%\n", bias*100.)
	return mh
}
