package pet

import "math"

const vonK = .41

// aerodynamicResistance (ra) [s/m]
func aerodynamicResistance(u float64) float64 {
	z, d, zo := 2., 0., .01 // [m]
	ra := math.Log((z - d) / zo)
	ra *= ra
	ra /= vonK * vonK * u
	return ra
}

// PenmanMonteith mass density flux m/s ~m³/m²/s
// Monteith, J.L. (1965) Evaporation and environment. Symposia of the Society for Experimental Biology 19: 205–224.
// see eq.10.25 in Hydrology in Practice pg.211
// Rn, G [W/m²]; t [°C]; p [Pa]; rh [0,1]; u [m/s]; rc [s/m]
func PenmanMonteith(Rn, G, t, p, rh, u, rc float64) float64 {
	m, g, l := slopeSaturationCurve(t), psychrometricConstant(t, p), latenHeatVapouration(t) // [Pa/°C],[Pa/°C],[MJ/kg]
	l *= 1.e6 * densityLiquidWater(t)                                                        // [W·s/m³]
	pa := densityDryAir(t)                                                                   // densityMoistAir(t, rh) [g/m³]
	de := vapourPressureDeficit(t, rh)                                                       // [Pa]
	ra := aerodynamicResistance(u)                                                           // [s/m]
	lE := (m*(Rn-G) + pa*ca*de/ra) / (m + g*(1.+rc/ra))                                      // [W/m²]
	return lE / l                                                                            // [m/s]
}
