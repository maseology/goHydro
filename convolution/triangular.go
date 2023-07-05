package convolution

import "math"

// Triangular similar to the HBV MAXBAS transfer function with the option of skewing the mode
// ref: Seibert, J. and J.J. McDonnell, 2010. Land-cover impacts on streamflow: a change-detection modelling approach that incorporates parameter uncertainty. Hydrological Sciences Journal 55(3), pp. 316-332.
// parameter Base is the trangular base and is in terms of number of timesteps (may not necessarily be discrete)
// parameter Skew represents a percentage along the triangular base; 50% represents a centered mode (i.e., equilateral triangle)
// ouput is in the form of percent effective runoff passing the calibration gauge for every discrete timestep
// see page 301 in Law, A.M., 2007. Simulation Modeling & Analysis, 4th ed. McGraw-Hill. 768 pp.
func Triangular(base, skew, lag float64) []float64 {
	if base < 0. || lag < 0. || skew < 0. || skew > 1. {
		panic("Triangular input error")
	}
	a, b, m := lag, base+lag, skew*base+lag
	bi := int(b) - 1
	if b-float64(bi) > 0. {
		bi++
	}
	ttf := make([]float64, bi, bi)
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
