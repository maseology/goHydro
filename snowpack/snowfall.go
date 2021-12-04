package snowpack

// SnowFallDensity assumes a linear relationship between temperature and snowfall density
// (correlation coefficient of 0.52 found in Judson and Doesken, 2000)
// Judson, A. and N. Doesken, 2000. Density of Freshly Fallen Snow in the Central Rocky Mountains. Bulletin of the American Meteorological Society, 81(7): 1577-1587.)
func SnowFallDensity(t float64) float64 {
	// const cdt = 5.5 // [kg/m³/°C] slope of density-temperature relationship
	if t > 0. {
		return den0
	}
	d := cdt*t + den0
	if d < denmin {
		return denmin
	}
	return d
}
