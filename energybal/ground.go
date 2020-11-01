package energybal

// GroundHeatFlux [W/m²]
// kg: thermal conductivity of soil [W/m/K] (pg.171 DeWalle & Rango)
// ts: soil temperature [°C] measured at depth zts
func GroundHeatFlux(tC, ts, zts, kg float64) float64 {
	// if snow present, tC>=0.
	return kg * (tC - ts) / zts
}
