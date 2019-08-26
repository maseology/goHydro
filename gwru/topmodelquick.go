package gwru

import (
	"fmt"
	"math"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

// TMQ is an optimized, distributed variation of the TOPMODEL struct
type TMQ struct {
	d, Qs     map[int]float64 // cell deficit relative to Dm; saturated lateral discharge (=w To tan(beta)/a)
	Dm, M, Ca float64         // mean deficit, discharge at Dm=0, Prameter m, catchment area
}

// Copy TMQ
func (t *TMQ) Copy() TMQ {
	return TMQ{
		d:  mmio.CopyMapif(t.d),
		Dm: t.Dm,
		M:  t.M,
		Ca: t.Ca,
	}
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
