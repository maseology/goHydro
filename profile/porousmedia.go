package profile

import (
	"github.com/maseology/goHydro/porousmedia"
)

// rPM is an alias for PorousMedium needed to add methods to the root struct for the Richards 1D solver.
type rPM struct {
	*porousmedia.PorousMedium
	cn float64 // Campbell (1974) shape parameter = 2+3/b
	b2 float64 // = 1+3/b
}

// newPM returns a new instance of the PorousMedium alias
func newPM(p *porousmedia.PorousMedium) *rPM {
	return &rPM{p, 2. + 3./p.B, 1. + 3./p.B} // Campbell (1974) shape parameter n=2+3/b; and (1-n)=-(1+3/b)
}

func (pm *rPM) dtdp(psi float64) float64 {
	if psi > pm.He {
		return 0.
	}
	return -pm.GetTheta(psi) / (pm.B * psi) // capacity (Campbell, 1974)
}

// vapour exchange
func (pm *rPM) GetKvap(q, theta float64) float64 {
	return (pm.Ts - theta) * mw * eta * rhoa * da * q / r / ts
}

// func (pm *rPM) GetSpecificHumidity(psi float64) float64 {
// 	return qp * math.Exp(mw*psi/r/ts)
// }

// func (pm *rPM) dqdp(q float64) float64 {
// 	return q * mw / r / ts
// }

// func (pm *rPM) qFromPsiTheta(psi, theta float64) float64 {
// 	return (pm.Ts - theta) * qp * math.Exp(mw*psi/r/ts) // specific humidity
// }

// func (pm *rPM) dqdp(q, psi, theta float64) float64 {
// 	return q * ((pm.Ts-theta)*mw/r/ts + theta/(pm.B*psi))
// }
