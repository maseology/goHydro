package pet

// const (
// 	alpha = .61
// 	beta  = -1.2e-4 // [m/d]
// )

// Oudin return daily potential evaporation [m/d]
// ref: L. Oudin, F. Hervieu, C. Michel, C. Perrin, V. Andreassian, F. Anctil and C. Loumagne, Which potential evapotranspiration input for a lumped rainfall-runoff model? Part 2—Towards a simple and efficient potential evapotranspiration model for rainfall-runoff modeling, Journal of Hydrology, 303, 2005, pp. 290–306.
// as shown in HMETS: Martel, J., Demeester, K., Brissette, F., Poulin, A., Arsenault, R., 2017. HMETS - a simple and efficient hydrology model for teaching hydrological modelling, flow forecasting and climate change impacts to civil engineering students. International Journal of Engineering Education 34, 1307–1316.
// kg: global sw radiation [MJ/m²/day]
// tm: daily mean temperature [°C]
func Oudin(kg, tm float64) float64 { // [m/d]
	if tm+5. > 0. {
		l := latenHeatVapouration(tm) * densityLiquidWater(tm) // convert MJ/kg to MJ/m³
		// return kg / l * (tm + 5.) / 100. // [mm/d]
		return kg / l * (tm + 5.) / 1e5
	}
	return 0.
}
