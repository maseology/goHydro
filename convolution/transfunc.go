package convolution

// TF is a general transfer function implimenter
type TF struct {
	SQ, QT []float64
}

// NewTriangularTF creates a new triangular weighted transfer function
// Triangular similar to the HBV MAXBAS transfer function with the option of skewing the mode
// ref: Seibert, J. and J.J. McDonnell, 2010. Land-cover impacts on streamflow: a change-detection modelling approach that incorporates parameter uncertainty. Hydrological Sciences Journal 55(3), pp. 316-332.
// parameter Base is the trangular base and is in terms of number of timesteps (may not necessarily be discrete)
// parameter Skew represents a percentage along the triangular base; 50% represents a centered mode (i.e., equilateral triangle)
// output is in the form of percent effective runoff passing the calibration gauge for every discrete timestep
func NewTriangularTF(base, skew, lag float64) TF {
	if base < 0. || skew < 0. || skew > 1. {
		panic("NewTTF input error")
	}
	a, b, m := lag, base+lag, skew*base+lag
	qt := Triangular(a, b, m) // MAXBAS: triangular weighted transfer function
	return TF{
		QT: qt,
		SQ: make([]float64, len(qt)+1), // delayed runoff
	}
}
