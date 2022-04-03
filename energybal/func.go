package energybal

const (
	emissivity = .97
	stefBoltz  = 5.67e-8 // [W/m²/K^4]
	oneseventh = 1. / 7.
)

func saturationVapourPressure(tC float64) float64 { // [Pa]
	// August-Roche-Magnus approximation (from pg.38 of Lu, N. and J.W. Godt, 2013. Hillslope Hydrology and Stability. Cambridge University Press. 437pp.)
	// for -30°C =< T =< 50°C
	return 610.49 * math.Exp(17.625*tC/(tC+243.04)) // [Pa=N/m²] R²=1
}