package profile

import "github.com/maseology/goHydro/porousmedia"

// Profile contains a set of layered PorousMedium struct used
// to describe a vertically layered soil profile
type Profile struct {
	P map[int]*porousmedia.PorousMedium
	D map[int]float64 // depths relative to top T
	T float64         // top of profile
}

// New constuctor for Profile
func (p *Profile) New(pm porousmedia.PorousMedium) {
	// single homogenous soil layer
	p.D = make(map[int]float64)
	p.P = make(map[int]*porousmedia.PorousMedium)
	p.P[1] = &pm
	p.T = ztop
	p.D[1] = pdpth
}

// GetPorousMedium returns the PorousMedium type from depth
func (p *Profile) GetPorousMedium(depth float64) *porousmedia.PorousMedium {
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
