package gwru

import (
	"fmt"
	"math"

	"github.com/maseology/mmaths"
)

// TMQ is an optimized variation of the TOPMODEL struct
type TMQ struct {
	d             map[int]float64 // cell deficit
	Dm, Qo, M, Ca float64         // mean deficit, discharge at Dm=0, Prameter m, catchment area
}

// Update state. input g: total basin average recharge per time step [m]
func (t *TMQ) Update(g float64) float64 {
	qb := t.Qo * math.Exp(-t.Dm/t.M) // gw discharge to streams [m/ts]
	t.Dm -= g                        // add recharge [m/ts]
	t.Dm += qb                       // remove baseflow discharge
	return qb                        // baseflow [m/ts]
}

// GetDi returns the local subsurface deficit (Di)
func (t *TMQ) GetDi(cid int) float64 {
	return t.Dm + t.d[cid]
}

// IsSame compares two TMQ structs
func (t *TMQ) IsSame(t1 *TMQ) (bool, string) {
	if t.M != t1.M {
		return false, "m"
	}
	if t.Qo != t1.Qo {
		return false, "Qo"
	}
	if t.Dm != t1.Dm {
		return false, "Dm"
	}
	if t.Ca != t1.Ca {
		return false, "ca"
	}

	c := 0
	for i, d := range t.d {
		dd := t1.d[i]
		rd := math.Abs(mmaths.RelativeDifference(d, dd))
		if rd > 1e-10 {
			c++
		}
	}
	if c > 0 {
		return false, fmt.Sprintf("t: %d", c)
	}

	return true, "nil"
}
