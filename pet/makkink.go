package pet

const (
	alpha = .61
	beta  = -1.2e-4 // [m/d]
)

// Makkink return daily potential evaporation [m/d]
// ref: Makkink, G.F., 1957. Ekzameno de la Formulo de Penman. Netherlands Journal of Agricultural Science 5:290--305.
// Kg: global sw radiation (MJ/m²)
// tm: daily mean temperature [°C]
// p: atmospheric pressure [Pa]
func Makkink(Kg, tm, p float64) float64 { // [m/d]
	if tm <= 0. {
		return 0.
	}
	d, g, l := slopeSaturationCurve(tm), psychrometricConstant(tm, p), latenHeatVapouration(tm)
	l *= densityLiquidWater(tm) // convert MJ/kg to MJ/m³
	return alpha*d*Kg/(d+g)/l + beta
}
