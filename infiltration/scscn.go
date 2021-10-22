package infiltration

import "math"

func Scscn(p, cn float64, amc int) float64 {
	// The Natural Resources Conservation Service (NRCS), formerly known as the Soil Conservation Service (SCS)
	// NOTE: Initial abstraction (Ia) losses should have been accounted for; therefore, the dormant season corrections
	//        have been maintained all-year, as initial losses (REF??) assumed during growing seasons have been accounted for
	if cn >= 100. {
		return 0.
	}
	// Adjust CN based on anticedent soil conditions:
	// AMC conversion factors from: Hawkins, R.H., A.T. Hjelmfelt, A.W. Zevenbergen, 1985. Runoff Probability, Storm Depth, and Curve Numbers. Journal of the Irrigation and Drainage Division, ASCE 111(4). pp.330-340.
	if amc == 1 { // dry AMCI
		cn /= .281 - 0.01281*cn
	} else if amc == 3 { // wet AMCIII
		cn /= .427 + 0.00573*cn
	}
	if cn == 0 {
		return p
	}
	// scn := 1000. / cn - 10. // inches
	scn := 25.4/cn - 0.254             // m
	return p - math.Pow(p, 2.)/(p+scn) // returns infiltration
}
