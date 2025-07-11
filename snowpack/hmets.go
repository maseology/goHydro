package snowpack

// HMETS snowmelt method
// Martel, J., Demeester, K., Brissette, F., Poulin, A., Arsenault, R., 2017. HMETS - a simple and efficient hydrology model for teaching hydrological modelling, flow forecasting and climate change impacts to civil engineering students. International Journal of Engineering Education 34, 1307–1316.
type HMETS struct {
	// snowpack
	ddfmin, ddfplus, tbm, kcum, fcmin, fcplus, ccum, tbf, kf, fe float64
}

func NewHMETS(ddfmin, ddfplus, tbm, kcum, fcmin, fcplus, ccum, tbf, kf, fe float64) HMETS {
	d := HMETS{
		ddfmin:  ddfmin,  // Minimum degree-day-factor [mm/°C/day]
		ddfplus: ddfplus, // Maximum degree-day-factor [mm/°C/day] (ddfmin + ddfplus = ddfmax)
		tbm:     tbm,     // Base melting temperature [°C]
		kcum:    kcum,    // Empirical parameter for the calculation of the degree-day-factor [mm–1]
		fcmin:   fcmin,   // Minimum fraction for the snowpack water retention capacity
		fcplus:  fcplus,  // Maximum fraction of the snowpack water retention capacity (fcmin + fcplus = fcmax)
		ccum:    ccum,    // Parameter for the calculation of water retention capacity [mm–1]
		tbf:     tbf,     // Base refreezing temperature [°C]
		kf:      kf,      // Degree-day factor for refreezing [mm/°C/day]
		fe:      fe,      // Empirical exponent for the freezing equation
	}
	return d
}

func (d *HMETS) Update(r, s, t float64) (melt, throughfall float64) {
	panic("todo")
	return -9999., -9999.
}
