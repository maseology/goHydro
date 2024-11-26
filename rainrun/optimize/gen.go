package optimize

import (
	"log"
	"math"

	rr "github.com/maseology/goHydro/rainrun"
	"github.com/maseology/goHydro/rainrun/sample"
)

func eval(m rr.Lumper) float64 { // evaluate model
	o := make([]float64, gfrc.Ndt)
	s := make([]float64, gfrc.Ndt)
	for i, v := range gfrc.D {
		y := v.Yield()
		_, r, _ := m.Update(y, v.Ep)
		o[i] = v.Q
		s[i] = r
	}
	return minimizer(o[365:], s[365:])
}

func genAtkinson(u []float64) float64 {
	var m rr.Lumper = &rr.Atkinson{}
	m.New(sample.Atkinson(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		log.Fatalf("Objective function error, u: %v\n", u)
	}
	return f
}

func genDawdyODonnell(u []float64) float64 {
	var m rr.Lumper = &rr.DawdyODonnell{}
	m.New(sample.DawdyODonnell(u, gfrc.Timestep)...)
	f := eval(m)
	if math.IsNaN(f) {
		log.Fatalf("Objective function error, u: %v\n", u)
	}
	return f
}

func genGR4J(u []float64) float64 {
	var m rr.Lumper = &rr.GR4J{}
	m.New(sample.GR4J(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		log.Fatalf("Objective function error, u: %v\n", u)
	}
	return f
}

// func genGR4J(u []float64) float64 {
// 	var m rr.Lumper = &rr.GR4J{}
// 	ss := sampler.NewSet(sample.GR4J()) //////////////////////////////////  TO FIX
// 	m.New(ss.Sample(u)...)
// 	f := eval(m)
// 	if math.IsNaN(f) {
// 		log.Fatalf("Objective function error, u: %v\n", u)
// 	}
// 	return f
// }

func genHBV(u []float64) float64 {
	var m rr.Lumper = &rr.HBV{}
	m.New(sample.HBV(u, gfrc.Timestep)...)
	f := eval(m)
	if math.IsNaN(f) {
		// log.Fatalf("Objective function error, u: %v\n", u)
		return 1000.
	}
	return f
}

func genManabeGW(u []float64) float64 {
	var m rr.Lumper = &rr.ManabeGW{}
	m.New(sample.ManabeGW(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		// log.Fatalf("Objective function error, u: %v\n", u)
		return 1000.
	}
	return f
}

func genMultiLayerCapacitance(u []float64) float64 {
	var m rr.Lumper = &rr.MultiLayerCapacitance{}
	m.New(sample.MultiLayerCapacitance(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		// log.Fatalf("Objective function error, u: %v\n", u)
		return 1000.
	}
	return f
}

func genQuinn(u []float64) float64 {
	var m rr.Lumper = &rr.Quinn{}
	m.New(sample.Quinn(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		log.Fatalf("Objective function error, u: %v\n", u)
	}
	return f
}

func genSIXPAR(u []float64) float64 {
	var m rr.Lumper = &rr.SIXPAR{}
	m.New(sample.SIXPAR(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		// log.Fatalf("Objective function error, u: %v\n", u)
		return 1000.
	}
	return f
}

func genSPLR(u []float64) float64 {
	var m rr.Lumper = &rr.SPLR{}
	m.New(sample.SPLR(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		log.Fatalf("Objective function error, u: %v\n", u)
	}
	return f
}

func genTank(u []float64) float64 {
	var m rr.Lumper = &rr.Tank{}
	m.New(sample.Tank(u)...)
	f := eval(m)
	if math.IsNaN(f) {
		log.Fatalf("Objective function error, u: %v\n", u)
	}
	return f
}
