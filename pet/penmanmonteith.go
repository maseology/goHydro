package pet

import (
	"math"
)

const vonK = .41

// aerodynamicResistance (ra) [s/m]
func aerodynamicResistance(u, t, rh, zo, q float64) float64 {
	// // ra := math.Log((z - d) / zo)
	// ra := math.Log(2. / zo)
	// ra *= ra
	// ra /= vonK * vonK * u
	// return ra

	// pg.241, eq.12.5
	va := (9e-6*t*t + 0.0088*t + 1.3281) * 1e-5 // kinematic viscosity of air at 1 atm m²/s (see air.xlsx)
	beta := func() float64 {
		// pg.52 Novak
		if u > 2.5 {
			return 4.62
		} else if u >= 1.5 {
			return 1.23
		}
		return .45
	}()

	us := u / math.Log(2./zo)
	lp := -math.Pow(us, 3.) / vonK / (9.81 / t) / (q * ca / densityMoistAir(t, rh)) // q: kinematic heat flux
	ra := zo*us/va + math.Log(2./zo) + beta*(2./lp)
	return ra / vonK / us
}

func windFunction(u, a, b float64) float64 {
	// ref: Penman (1948)
	// return math.Pow(a, b)
	return a + b*u
}

// PenmanMonteith mass density flux m/s ~m³/m²/s
// Monteith, J.L. (1965) Evaporation and environment. Symposia of the Society for Experimental Biology 19: 205–224.
// see eq.10.25 in Hydrology in Practice pg.211
// Rn [W/m²]; t [°C]; p [Pa]; rh [0,1]; u [m/s]; rc [s/m]
func PenmanMonteith(Rn, t, p, rh, u, rc, a, b float64) (float64, float64) {
	m, g, l := slopeSaturationCurve(t), psychrometricConstant(t, p), latenHeatVapouration(t) // [Pa/°C],[Pa/°C],[MJ/kg]
	pa, pw := densityDryAir(t)/1000., densityLiquidWater(t)                                  // [kg/m³]      // pa := densityMoistAir(t, rh)/1000.
	l *= 1.e6 * pw                                                                           // [W·s/m³]

	de := vapourPressureDeficit(t, rh) // [Pa]
	// ra := aerodynamicResistance(u, t, rh, zo, 1111111.) // [s/m]
	Dv := windFunction(u, a, b) // [m/s]
	o := m / (m + g)            // [--]
	H := o * Rn / l             // [m/s]
	// Ea := (1. - o) * pa * mwr / p * de / ra / pw        // [m/s]
	Ea := (1. - o) * pa * mwr / p * de * Dv / pw // [m/s]

	return H, Ea // [m/s]

	// m, g, l := slopeSaturationCurve(t), psychrometricConstant(t, p), latenHeatVapouration(t) // [Pa/°C],[Pa/°C],[MJ/kg]
	// l *= 1.e6 * densityLiquidWater(t)                                                        // [W·s/m³]
	// pa := densityDryAir(t)                                                                   // densityMoistAir(t, rh) [g/m³]
	// de := vapourPressureDeficit(t, rh)                                                       // [Pa]
	// ra := aerodynamicResistance(u)                                                           // [s/m]
	// lE := (m*(Rn-G) + pa*ca*de/ra) / (m + g*(1.+rc/ra))                                      // [W/m²]
	// return lE / l, Ea                                                                        // [m/s]
}
