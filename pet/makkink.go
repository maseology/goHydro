package pet

// const (
// 	alpha = .61
// 	beta  = -1.2e-4 // [m/d]
// )

// Makkink return daily potential evaporation [m/d]
// ref: Makkink, G.F., 1957. Ekzameno de la Formulo de Penman. Netherlands Journal of Agricultural Science 5:290--305.
// Kg: global sw radiation [MJ/m²]
// tm: daily mean temperature [°C]
// p: atmospheric pressure [Pa]
// alpha, beta are coefficients (0.61 and -1.2e-4 m/d, respectively)
func Makkink(Kg, tm, p, alpha, beta float64) float64 { // [m/d]
	if tm <= 0. {
		return 0.
	}
	d, g := slopeSaturationCurve(tm), psychrometricConstant(tm, p)
	l := latenHeatVapouration(tm) * densityLiquidWater(tm) // convert MJ/kg to MJ/m³
	ep := alpha*Kg*d/(d+g)/l + beta
	if ep < 0. {
		ep = 0.
	}
	return ep
}
