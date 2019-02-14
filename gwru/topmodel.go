package gwru

import (
	"math"
)

// TOPMODEL struct
type TOPMODEL struct {
	ti, Di       map[int]float64
	g, qo, m, ca float64
}

// Update state. input g: total basin average recharge per time step [m]
// returns baseflow
func (t *TOPMODEL) Update(g float64) float64 {
	// returns baseflow
	dm := 0.
	for _, v := range t.Di {
		dm += v
	}
	dm /= float64(len(t.Di))
	dm -= g

	qb := t.qo * math.Exp(-dm/t.m)
	dm += qb / t.ca

	t.updateDeficits(dm)
	return qb
}

// Storage returns total current storage
func (t *TOPMODEL) Storage() float64 {
	return 1.
}

// TopographicIndex returns total current storage
func (t *TOPMODEL) TopographicIndex() map[int]float64 {
	return t.ti
}

func (t *TOPMODEL) updateDeficits(dm float64) {
	for i, v := range t.ti {
		t.Di[i] = dm + t.m*(t.g-v)
	}
}
