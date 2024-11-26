package rainrun

import "math"

// Tank model
// ref: Sugawara, M. (1995). Tank model. In: V.P. Singh (Ed.), Computer models of watershed hydrology. Water Resources Publications, Highlands Ranch, Colorado.
type Tank struct{ z11, z12, a11, a12, b1, h1, z2, a2, b2, h2, z3, a3, b3, h3, a4, h4 float64 }

// New tank model
// [z11, z12, z2, z3, a11, a12, a2, a3, a4, b1, b2, b3]
func (t *Tank) New(p ...float64) {
	for i := range 8 {
		if p[i+4] < 0. || p[i+4] > 1. {
			println(p[4:])
			panic("Tank input error")
		}
	}
	t.z11 = p[0]
	t.z12 = p[1]
	t.z2 = p[2]
	t.z3 = p[3]
	t.a11 = p[4]
	t.a12 = p[5]
	t.a2 = p[6]
	t.a3 = p[7]
	t.a4 = p[8]
	t.b1 = p[9]
	t.b2 = p[10]
	t.b3 = p[11]
}

// Storage returns total storage
func (t *Tank) Storage() float64 {
	return t.h1 + t.h2 + t.h3 + t.h4
}

// Update state for daily inputs
func (t *Tank) Update(p, ep float64) (float64, float64, float64) {
	var a float64
	t.h1 += p
	if ep > t.h1 {
		a = t.h1
		t.h1 = 0.
	} else {
		a = ep
		t.h1 -= ep
	}
	q1 := math.Max(0., t.h1-t.z11)*t.a11 + math.Max(0., t.h1-t.z12)*t.a12
	inf := t.b1 * t.h1
	t.h1 -= q1 + inf

	t.h2 += inf
	q2 := math.Max(0., t.h2-t.z2) * t.a2
	perc := t.b2 * t.h2
	t.h2 -= q2 + perc

	t.h3 += perc
	q3 := math.Max(0., t.h3-t.z3) * t.a3
	deepperc := t.b3 * t.h3
	t.h3 -= q3 + deepperc

	t.h4 += deepperc
	q4 := t.h4 * t.a4
	t.h4 -= q4

	return a, q1 + q2 + q3 + q4, deepperc
}
