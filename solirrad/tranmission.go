package solirrad

import . "math"

// ApplyAtmosphericTransmissionCoefficient applies a correction factor to account for losses due to atmospheric scattering
// see: Buffo, Fritschen, Murphy, 1972. USDA: Direct Solar Radiation on Various Slopes From 0 to 60 Degrees North Latitude
// See also pg. 7 in Budyko: Climate and Life, eq. 1.16; pg. 345 in Oke: Boundary Layer Climates
// coef = 0.6 smoggy air to 0.9 clean and clear, typical values 0.84 (see Oke and references therein)
// Note: altitude angle (A) is complementary to zenith (i.e., A = 90-Z; or sin(A) = cos(Z))
func (si *SolIrad) ApplyAtmosphericTransmissionCoefficient(coef float64) {
	if si.Lat > 60.*Pi { // upper limit acording to Budyko - offers other methods (see below)
		panic("upper latitude limit for applying the Atmospheric Transmission Coefficient")
	}
	for i := 0; i <= 365; i++ {
		si.psi[i] *= Pow(coef, 1./Cos(si.zeff[i]))
		si.psih[i] *= Pow(coef, 1./Cos(si.zeffh[i]))
	}
}

// NetSWfromPotential is an empirical conversion from incoming potential radiantion to incomming SW radiation arriving at the surface
// see pg 151 in DeWalle & Rango; Also attributed to Linacre (1992), but the above form is preferred as sunshine hours are easiest to determine
func (si *SolIrad) NetSWfromPotential(CloudCoverFraction float64, DayOfYear int) float64 {
	return (0.85 - 0.47*CloudCoverFraction) * si.psi[DayOfYear-1]
}

// GlobalFromPotential is an empirical conversion from incoming potential radiantion to global radiation arriving at the surface
// see pg 151 in DeWalle & Rango; attributed to Bristow and Campbell (1984)
// ref: Bristow, K.L. and G.S. Campbell, 1984. On the relationship between incoming solar radiation and daily maximum and minimum temperature. Agricultural and Forest Meteorology 31(2):159--166.
func (si *SolIrad) GlobalFromPotential(tx, tn, a, b, c float64, DayOfYear int) float64 {
	delT := tx - tn // (tn0+tn1)/2.
	return si.psi[DayOfYear-1] * a * (1. - Exp(-b*Pow(delT, c)))
}
