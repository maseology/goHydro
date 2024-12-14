package hru

import "math"

// Tank stacked reservoir with infinite outlets
// modified from: Sugawara, M. (1995). Tank model. In: V.P. Singh (Ed.), Computer models of watershed hydrology. Water Resources Publications, Highlands Ranch, Colorado.
type Tank struct{ Dz, A, Sto float64 }

// Overflow : updates state. p is an net addition and function returns excess (+).
// If p<0 and |p|>sto, function returns remainder of sink (-).
func (t *Tank) Overflow(p float64) float64 {
	t.Sto += p
	if t.Sto < 0. {
		d := t.Sto
		t.Sto = 0.
		return d
	} else {
		q := 0.
		// sto0 := t.Sto
		if t.Sto > t.Dz {
			n := math.Floor(t.Sto / t.Dz)
			// q = (n-1)*t.Dz*t.A + (t.Sto-n*t.Dz)*t.A
			// cc := 0
			for i := n; i > 0; i-- {
				qi := (t.Sto - i*t.Dz) * t.A
				q += qi
				t.Sto -= qi
				// if math.IsNaN(t.Sto) {
				// 	print("")
				// }
				// cc++
				// if cc > 100000 {
				// 	print("")
				// }
			}
		}
		// _ = sto0
		// g := t.Sto * t.B
		// t.Sto -= g
		// t.Sto -= q
		return q
	}
}
