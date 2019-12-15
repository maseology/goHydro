package lia // VERSION 2

import "math"

func (d *Domain) setDT() {
	dmax := dmin //-math.MaxFloat64 // the maximum water depth over the domain
	for _, n := range d.ns {
		if n.h-n.z > dmax {
			dmax = n.h - n.z
		}
	}
	d.dt = d.Alpha * d.dx / math.Sqrt(g*dmax) // eq.12
	d.tcum += d.dt
}
