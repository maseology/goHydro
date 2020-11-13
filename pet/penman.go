package pet

import (
	"math"
)

func windFunction(u, a, b float64) float64 {
	// ref: Penman (1948)
	// return math.Pow(a, b)
	return a + b*u
}

// Penman mass density flux m/s ~m³/m²/s
// Penman (1948)
// Rn [W/m²]; t [°C]; p [Pa]; rh [0,1]; u [m/s]
func Penman(Rn, t, p, rh, u, a, b float64) (float64, float64) {
	m, g, l := slopeSaturationCurve(t), psychrometricConstant(t, p), latenHeatVapouration(t) // [Pa/°C],[Pa/°C],[MJ/kg]
	pa, pw := densityDryAir(t)/1000., densityLiquidWater(t)                                  // [kg/m³]      // pa := densityMoistAir(t, rh)/1000.
	l *= 1.e6 * pw                                                                           // [W·s/m³]

	de := vapourPressureDeficit(t, rh)           // [Pa]
	Dv := windFunction(u, a, b)                  // [m/s]
	o := m / (m + g)                             // [--]
	H := o * Rn / l                              // [m/s]
	Ea := (1. - o) * pa * mwr / p * de * Dv / pw // [m/s]

	return H, Ea // [m/s]
}

// PenmanWind mass density flux m/s ~m³/m²/s
// Penman (1948) based solely on the wind function
// see Novak 182
func PenmanWind(t, p, rh, u, a, b float64) float64 {
	// pa, pw := densityDryAir(t)/1000., densityLiquidWater(t) // [kg/m³],[kg/m³]       // pa := densityMoistAir(t, rh)/1000.
	// de := vapourPressureDeficit(t, rh)                      // [Pa]
	// return a * mwr * pa / pw * de / p * math.Pow(u, b)      // [m/s]

	de := vapourPressureDeficit(t, rh)       // [Pa]
	return a * de * math.Pow(u, b) * 7.46e-9 // [m/s]
}
