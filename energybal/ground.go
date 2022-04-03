package energybal

// GroundHeatFlux [W/m²]
// kg: thermal conductivity of soil [W/m/K] (pg.171 DeWalle & Rango)  .2-2. W/m/K (Oke, 1987)
// ts: soil temperature [°C] measured at depth zts
func GroundHeatFlux(Tair, ts, zts, kg float64) float64 {
	// if snow present, Tair>=0.
	return kg * (Tair - ts) / zts
}
