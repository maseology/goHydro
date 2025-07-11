package convolution

import "math"

// NewDiffusiveWaveConvolution build a Convolution based on an analytical solution to the diffusive wave equation at some reference discharge (qRef [m3/s])
//
// bedslope [-]; reachlength [m]
func NewDiffusiveWaveConvolution(rc *RatingCurve, qRef, reachlength, bedslope, tsec float64) *Convolution {
	// (copied from Raven source code)

	cRef := getCelerity(rc, qRef)                     // [m/s]
	diffusivity := getDiffusivity(rc, qRef, bedslope) // [m2/s]
	traveltime := reachlength / cRef                  // [s]
	n := int(math.Ceil(2*traveltime/tsec)) + 2

	ws, qs := make([]float64, n), make([]float64, n)
	for i := range n {
		qs[i] = qRef // initialize
	}
	if traveltime < tsec { //very sharp ADR CDF or reach length==0 -override
		ws[0] = 1. - traveltime/tsec
		ws[1] = traveltime / tsec
		for i := 2; i < n; i++ {
			ws[i] = 0.
		}
	} else {
		s := 0.
		for i := range n - 1 {
			ws[i] = math.Max(ogataBanks(float64(i)*tsec, diffusivity, reachlength, cRef)-s, 0.)
			s += ws[i]
		}
		ws[n-1] = 0. // must truncate infinite distrib.
		for i := range n - 1 {
			ws[i] /= s // normalize
		}
	}

	return &Convolution{ws, qs}
}

// Calculates cumulative kinematic wave solution distribution
// t time [d]; L reach length [m]; v celerity [m/d]; D diffusivity [m2/d]
// int_0^time L/2/t^(3/2)/sqrt(pi*D)*exp(-(v*t-L)^2/(4*D*t)) dt
// extreme case: (D->0): =1 for v*t<L, 0 otherwise
// (function taken from Raven source code)
func ogataBanks(t, D, L, v float64) float64 {
	//Analytical solution by Ogata Banks, 1969, eq.13
	F := 0.
	if t <= 0 {
		return 0.
	}
	if L < 500*(D/v) {
		F = 0.5 * (math.Exp((v*L)/D) * math.Erfc((L+v*t)/math.Sqrt(4.*D*t)))
	}
	F += 0.5 * (math.Erfc((L - v*t) / math.Sqrt(4.*D*t)))
	return F
}

func getCelerity(rc *RatingCurve, qref float64) float64 {
	if qref < rc.Q[0] {
		return (5. / 3.) * rc.Q[0] / rc.A[0] // pg.279 Ponce (1989) //[m/s]
	}
	for i, q := range rc.Q {
		if i > 0 && qref < q {
			return (rc.Q[i] - rc.Q[i-1]) / (rc.A[i] - rc.A[i-1]) // dQ/dA
		}
	}
	panic("should increase rating curve max depth")
	return (5. / 3.) * qref / rc.A[len(rc.A)-1]
}

func getDiffusivity(rc *RatingCurve, qref, bedslope float64) float64 {
	tw := func() float64 {
		if qref < rc.Q[0] {
			return rc.W[0]
		}
		for i, q := range rc.Q {
			if i > 0 && qref < q {
				return rc.W[i]
			}
		}
		return rc.W[len(rc.W)-1]
	}()
	return qref / 2. / tw / bedslope // pg.290 Ponce (1989) //[m2/s]
}
