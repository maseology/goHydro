package swat

import "math"

func (m *HRU) evap(ep float64) (esum float64) {
	// et := 0.                            // Penman-Monteith is not used; therefore no transpiration
	// ep0 := ep                           // while using SCSCN, canopy extraction is implied (pg.124)
	// es := ep0 * m.cov                   // maximum potential soil evaporation/sublimation (pg.135)
	// eps := math.Min(es, es*ep0/(es+et)) // maximum potential soil evaporation/sublimation adjusted for plant use (pg.136)
	// esub := 0.                          // sublimation has already been accounted for (pg.137)
	// epps := eps - esub                  // maximum potential soil water evaporation adjusted for plant use (pg.137) [mm]
	esum = 0.
	epps := ep * m.cov // reduced processes from above
	esoilz := make([]float64, nsl+1)
	for i := 0; i <= nsl; i++ {
		z := float64(i) * lythick                              // depth [mm]
		esoilz[i] = epps * z / (z + math.Exp(2.374-0.00713*z)) // evaporative demand at depth z (pg.137)
	}
	for i := 0; i < nsl; i++ {
		esoil := esoilz[i+1] - esoilz[i]*m.esco // evaporative demand at for layer i (pg.137)
		if m.sz[i].sw < m.sz[i].fc {
			esoil *= math.Exp(2.5 * (m.sz[i].sw - m.sz[i].fc) / (m.sz[i].fc - m.sz[i].wp))
		}
		eppsoil := math.Max(0., math.Min(esoil, 0.8*(m.sz[i].sw-m.sz[i].wp))) // amount of water removed from layer by evaporation (pg.140) [mm]
		m.sz[i].sw -= eppsoil
		esum += eppsoil
	}
	return
}
