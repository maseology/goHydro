package gwru

import (
	"math"
)

// TMQ is an optimized variation of the TOPMODEL struct
type TMQ struct {
	t             map[int]float64
	Dm, qo, m, ca float64
}

// Update state. input g: total basin average recharge per time step [m]
// returns baseflow [m³/ts]
func (t *TMQ) Update(g float64) float64 {
	t.Dm -= g // recharge [m/ts]
	qb := t.qo * math.Exp(-t.Dm/t.m)
	t.Dm += qb / t.ca // gw discharge to streams [m³/ts]
	return qb
}

// GetDi returns the local subsurface deficit (Di)
func (t *TMQ) GetDi(cid int) float64 {
	return t.Dm + t.t[cid]
}
