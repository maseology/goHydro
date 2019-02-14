package gwru

import "log"

// Linear reservoir struct
type Linear struct {
	s, k float64
}

// New constructor
func (l *Linear) New(k float64) {
	if k < 0. || k > 1. {
		log.Panicf("Linear GW reservoir error: assigned k = %v", k)
	}
	l.s = 0.
	l.k = k
}

// Update state. input g: total basin average recharge per time step [m]
// returns baseflow
func (l *Linear) Update(g float64) float64 {
	l.s += g
	slast := l.s
	l.s *= l.k
	return slast - l.s
}

// Storage returns total current storage
func (l *Linear) Storage() float64 {
	return l.s
}
