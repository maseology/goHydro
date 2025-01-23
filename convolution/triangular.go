package convolution

import "math"

// Triangular density distribution f(x)
// [a,b] mode m
// see page 301 in Law, A.M., 2007. Simulation Modeling & Analysis, 4th ed. McGraw-Hill. 768 pp.
func Triangular(a, b, m float64) []float64 {
	bi := int(b)
	if b-float64(bi) > .5 {
		bi++
	}
	ttf := make([]float64, bi)
	if bi == 1 {
		ttf[0] = 1.
	} else {
		s := 0.
		for i := range bi {
			x := float64(i) + .5 // midpoint
			if x >= a && x <= m {
				ttf[i] = 2 * (x - a) / (b - a) / (m - a)
			} else if x > m && x <= b {
				ttf[i] = 2 * (b - x) / (b - a) / (b - m)
			} else {
				ttf[i] = 0.
			}
			s += ttf[i]
		}
		if s != 1. {
			for i := 0; i < bi; i++ {
				ttf[i] /= s
			}
		}
	}
	return ttf
}

// TriangularCDF cumulative distribution F(x)
func TriangularCDF(a, b, m float64) []float64 {
	bi := int(b)
	if b-float64(bi) > 0. {
		bi++
	}
	ttf := make([]float64, bi)
	if bi == 1 {
		ttf[0] = 1.
	} else {
		var s float64
		for i := 0; i < bi; i++ {
			x := float64(i)
			if x > b {
				ttf[i] = 1.
			} else if x > m { // after mode
				ttf[i] = 1. - math.Pow(b-x, 2.)/(b-a)/(b-m)
			} else if x < a {
				ttf[i] = 0.
			} else { // prior to mode a<=x<=m
				ttf[i] = math.Pow(x-a, 2.) / (b - a) / (m - a)
			}
			s += ttf[i]
		}
		if s != 1. {
			for i := 0; i < bi; i++ {
				ttf[i] /= s
			}
		}
	}
	return ttf
}
