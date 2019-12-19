package lia // VERSION 2

import (
	"math"
)

func (d *Domain) setDT() {
	dmax := dmin //-math.MaxFloat64 // the maximum water depth over the domain
	for _, n := range d.ns {
		if n.h-n.z > dmax {
			dmax = n.h - n.z
		}
	}
	// for _, s := range d.WKR {
	// 	hx := math.Max(s.nb.h, s.nf.h) - s.zx
	// 	if hx > dmax {
	// 		dmax = hx
	// 	}
	// }
	// NOTE: could collect all cells with d < dmin, and not analyse them
	d.dt = d.Alpha * d.dx / math.Sqrt(g*dmax) // eq.12
	d.tcum += d.dt
}
