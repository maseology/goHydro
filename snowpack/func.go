package snowpack

// SnowFallDensity assumes a linear relationship between temperature and snowfall density
// (correlation coefficient of 0.52 found in Judson and Doesken, 2000)
func SnowFallDensity(t float64) float64 {
	// const cdt = 5.5 // [kg/m³/°C] slope of density-temperature relationship
	if t > 0. {
		return den0
	}
	d := cdt + den0
	if d < denmin {
		return denmin
	}
	return d
}
