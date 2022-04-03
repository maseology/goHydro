package energybal

import "math"

// NetLW returns [W/m²] L* based on Oke (1987)
// function of temperature, humidity, cloud cover fraction and empirical parameter b [0,1]
func NetLW(tC, rh, n, c float64) float64 {
	// from pg:373-375 Oke, 1987. Boundary Layer Climates 2nd ed.
	// also pg. 233 Novak;
	ea := saturationVapourPressure(tC) / 100.                                  // vapour pressure [mb] (=100Pa)
	eao := .575 * math.Pow(ea, oneseventh)                                     // Brutseart (1975) atmospheric emissivity with cloudless skies
	ccf := math.Floor(11.*n) / 10.                                             // cloud cover fraction "fraction of sky covered with cloud" (Oke, pg.373) vs. n/N "ratio of actual/possible hours of sunshine" (Penman, 1948)
	return stefBoltz * math.Pow(273.16+tC, 4.) * (eao - 1.) * (1. - c*ccf*ccf) // L* = f(T)f(ea)f(n)
}
