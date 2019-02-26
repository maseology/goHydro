package gwru

import (
	"math"
)

// TOPMODEL struct
type TOPMODEL struct {
	Di, ti           map[int]float64
	Dm, g, qo, m, ca float64
}

// Update state. input g: total basin average recharge per time step [m]
// returns baseflow [m³/ts]
func (t *TOPMODEL) Update(g float64) float64 {
	// returns baseflow
	t.Dm = 0.
	for _, v := range t.Di {
		t.Dm += v
	}
	t.Dm /= float64(len(t.Di))
	t.Dm -= g // recharge [m/ts]

	qb := t.qo * math.Exp(-t.Dm/t.m)
	t.Dm += qb / t.ca // gw discharge to streams [m³/ts]

	t.updateDeficits()
	return qb
}

// Storage returns total current storage
func (t *TOPMODEL) Storage() float64 {
	return t.Dm // relative to full saturation [m]
}

// TopographicIndex returns total current storage
func (t *TOPMODEL) TopographicIndex() map[int]float64 {
	return t.ti
}

func (t *TOPMODEL) updateDeficits() {
	for i, v := range t.ti {
		t.Di[i] = t.Dm + t.m*(t.g-v)
	}
}
