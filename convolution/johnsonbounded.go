package convolution

import "math"

// JohnsonBounded density distribution f(x)
// [a,b], shape parameters a1~[-inf,inf], a2>0
// see page 297 in Law, A.M., 2007. Simulation Modeling & Analysis, 4th ed. McGraw-Hill. 768 pp.
//
func JohnsonBounded(a1, a2, a, b float64) []float64 {
	if a2 < 0 {
		panic("JohnsonBounded input error")
	}
	bi := int(b)
	if b-float64(bi) > 0. {
		bi++
	}
	ttf := make([]float64, bi)
	if bi == 1 {
		ttf[0] = 1.
	} else {
		s := 0.
		for i := range bi {
			x := float64(i) + .5 // midpoint
			if x > a && x < b {
				ttf[i] = a2 * (b - a) / (x - a) / (b - x) / math.Sqrt(2*math.Pi)
				ttf[i] *= math.Exp(-.5 * math.Pow(a1+a2*math.Log((x-a)/(b-x)), 2))
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

// // JohnsonBoundedCDF cumulative distribution F(x)
// func JohnsonBoundedCDF(a, b, m float64) []float64 {
//  panic("TODO")
// 	bi := int(b)
// 	if b-float64(bi) > 0. {
// 		bi++
// 	}
// 	ttf := make([]float64, bi)
// 	if bi == 1 {
// 		ttf[0] = 1.
// 	} else {
// 		var s float64
// 		for i := 0; i < bi; i++ {
// 			x := float64(i)
// 			if x > b {
// 				ttf[i] = 1.
// 			} else if x > m { // after mode
// 				ttf[i] = 1. - math.Pow(b-x, 2.)/(b-a)/(b-m)
// 			} else if x < a {
// 				ttf[i] = 0.
// 			} else { // prior to mode a<=x<=m
// 				ttf[i] = math.Pow(x-a, 2.) / (b - a) / (m - a)
// 			}
// 			s += ttf[i]
// 		}
// 		if s != 1. {
// 			for i := 0; i < bi; i++ {
// 				ttf[i] /= s
// 			}
// 		}
// 	}
// 	return ttf
// }
