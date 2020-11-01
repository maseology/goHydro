package energybal

import "math"

const (
	emissivity = .97
	stefBoltz  = 5.67e-8 // [W/m²/K^4]
	oneseventh = 1. / 7.
)

func saturationVapourPressure(tC float64) float64 { // [Pa]
	// August-Roche-Magnus approximation (from pg.38 of Lu, N. and J.W. Godt, 2013. Hillslope Hydrology and Stability. Cambridge University Press. 437pp.)
	// for -30°C =< T =< 50°C
	return 610.49 * math.Exp(17.625*tC/(tC+243.04)) // [Pa=N/m²] R²=1
}

// NetLW returns [W/m²] L* based on Oke (1987)
// function of temperature, humidity, cloud cover fraction and empirical parameter b [0,1]
func NetLW(tC, rh, ccf, b float64) float64 {
	// from pg:373-375 Oke, 1987. Boundary Layer Climates 2nd ed.
	// also pg. 233 Novak;
	ea := saturationVapourPressure(tC) / 100.                              // vapour pressure [mb] (=100Pa)
	eao := .575 * math.Pow(ea, oneseventh)                                 // Brutseart (1975) atmospheric emissivity with cloudless skies
	return stefBoltz * math.Pow(273.16+tC, 4.) * (eao - 1.) * (1. - b*ccf) // L* = f(T)f(ea)f(ccf)
}
