package hechms

type Params struct {
	Fia, Fcn             float64 // CN global multipliers
	Cp, Ct               float64 // Snyder
	Q0, Kbf, RatioToPeak float64 // Recession
	Krch, Xrch           float64 // storage coeffient, muskingum weighting factor
}
