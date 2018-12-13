package profile

import (
	. "github.com/maseology/goHydro/porousmedia"
)

// Profile contains a set of layered PorousMedium struct used
// to describe a vertically layered soil profile
type Profile struct {
	P map[int]*PorousMedium
	D map[int]float64 // depths relative to top T
	T float64         // top of profile
}

// New : constuctor for Profile
func (p *Profile) New() {
	// single homogenous soil layer
	p.D = make(map[int]float64)
	p.P = make(map[int]*PorousMedium)
	var pm PorousMedium
	pm.New()
	p.P[1] = &pm
	p.T = 0.0
	p.D[1] = 1.0
}

// GetPorousMedium returns the PorousMedium type from depth
func (p *Profile) GetPorousMedium(depth float64) *PorousMedium {
	if len(p.P) == 1 {
		return p.P[1]
	}
	for k := 1; k <= len(p.D); k++ {
		if k == 1 {
			if depth < p.D[1] {
				return p.P[1]
			}
		} else {
			if depth < p.D[k] {
				return p.P[k]
			}
		}
	}
	if depth == p.D[len(p.D)] {
		return p.P[len(p.D)]
	}
	return nil
}
