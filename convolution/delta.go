package convolution

import "math"

func DiracDelta(lag, ts float64) ([]float64, int) {
	if math.Mod(lag, ts) == 0 {
		lagm := int(lag / ts)
		o := make([]float64, lagm+1)
		o[lagm] = 1.
		return o, lagm
	}
	remain := math.Mod(lag, ts)
	lagm := int(math.Floor(lag / ts))
	o := make([]float64, lagm+2)
	o[lagm] = 1. - remain/ts
	o[lagm+1] = remain / ts
	return o, lagm + 1
}
