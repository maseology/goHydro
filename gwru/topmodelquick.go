package gwru

import (
	"fmt"
	"math"
	"log"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

const (
	omega   = 1. // sinuosity
)

// TMQ is an optimized, distributed variation of the TOPMODEL struct
type TMQ struct {
	D, Qs     map[int]float64 // cell deficit relative to Dm; saturated lateral discharge (=w To tan(beta)/a)
	Dm, M, Ca float64         // mean deficit, discharge at Dm=0, Prameter m, catchment area
}

// Copy TMQ
func (t *TMQ) Copy() TMQ {
	return TMQ{
		D:  mmio.CopyMapif(t.D),
		Dm: t.Dm,
		M:  t.M,
		Ca: t.Ca,
	}
}

// RelTi returns the topographical index relative to gamma (g-Ti)
func (t *TMQ) RelTi() map[int]float64 {
	out := make(map[int]float64, len(t.D))
	for k, v := range t.D {
		out[k] = v / t.M
	}
	return out
}

// // Clone creates a deep copy of TMQ, while changing recession coefficient m
// func (t *TMQ) Clone(m float64) TMQ {
// 	dnew := make(map[int]float64, len(t.d))
// 	for i, v := range t.d {
// 		dnew[i] = m * v / t.M
// 	}
// 	return TMQ{
// 		d:  dnew,
// 		Dm: 0.,
// 		M:  m,
// 		Ca: t.Ca,
// 	}
// }

// GetDi returns the local subsurface deficit (Di)
func (t *TMQ) GetDi(cid int) float64 {
	if d, ok := t.D[cid]; ok {
		return t.Dm + d
	}
	log.Fatalf(" TMQ.GetDi error: cell id %d not found in model\n", cid)
	return t.Dm 
}

// // UpdateLumped state. input g: total basin average recharge per time step [m]
// func (t *TMQ) UpdateLumped(g float64) float64 {
// 	qb := t.Qo * math.Exp(-t.Dm/t.M) // gw discharge to streams [m/ts]
// 	t.Dm -= g                        // add recharge [m/ts]
// 	t.Dm += qb                       // remove baseflow discharge
// 	return qb                        // baseflow [m/ts]
// }

// IsSame compares two TMQ structs
func (t *TMQ) IsSame(t1 *TMQ) (bool, string) {
	if t.M != t1.M {
		return false, "m"
	}
	if t.Dm != t1.Dm {
		return false, "Dm"
	}
	if t.Ca != t1.Ca {
		return false, "ca"
	}

	c := 0
	for i, d := range t.D {
		dd := t1.D[i]
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
