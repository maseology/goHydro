package pet

import (
	"math"
)

const vonK = .41

// aerodynamicResistance (ra) [s/m]
func aerodynamicResistance(u, zo float64) float64 {
	// ra := math.Log((z - d) / zo)
	ra := math.Log(2. / zo)
	ra *= ra
	ra /= vonK * vonK * u
	return ra
}

// PenmanMonteith mass density flux m/s ~m³/m²/s
// Monteith, J.L. (1965) Evaporation and environment. Symposia of the Society for Experimental Biology 19: 205–224.
// see eq.10.25 in Hydrology in Practice pg.211
// Rn, G [W/m²]; t [°C]; p [Pa]; rh [0,1]; u [m/s]; rc [s/m]
func PenmanMonteith(Rn, t, p, rh, u, rc, zo float64) (float64, float64) {
	m, g, l := slopeSaturationCurve(t), psychrometricConstant(t, p), latenHeatVapouration(t) // [Pa/°C],[Pa/°C],[MJ/kg]
	pa, pw := densityDryAir(t)/1000., densityLiquidWater(t)                                  // [kg/m³]      // pa := densityMoistAir(t, rh)/1000.
	l *= 1.e6 * pw                                                                           // [W·s/m³]

	de := vapourPressureDeficit(t, rh)           // [Pa]
	ra := aerodynamicResistance(u, zo)           // [s/m]
	o := m / (m + g)                             // [--]
	H := o * Rn / l                              // [m/s]
	Ea := (1. - o) * pa * mwr / p * de / ra / pw // [m/s]

	if H < 0 {
		print()
	}
	if Ea < 0 {
		print()
	}

	return H, Ea // [m/s]

	// m, g, l := slopeSaturationCurve(t), psychrometricConstant(t, p), latenHeatVapouration(t) // [Pa/°C],[Pa/°C],[MJ/kg]
	// l *= 1.e6 * densityLiquidWater(t)                                                        // [W·s/m³]
	// pa := densityDryAir(t)                                                                   // densityMoistAir(t, rh) [g/m³]
	// de := vapourPressureDeficit(t, rh)                                                       // [Pa]
	// ra := aerodynamicResistance(u)                                                           // [s/m]
	// lE := (m*(Rn-G) + pa*ca*de/ra) / (m + g*(1.+rc/ra))                                      // [W/m²]
	// return lE / l, Ea                                                                        // [m/s]
}
