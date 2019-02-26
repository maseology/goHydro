package swat

import "math"

// Update state (all in [mm])
func (bsn *SubBasin) Update(p, ep float64) (r, i, a, g, b float64) {
	r, i, a, g = 0., 0., 0., 0.
	for _, m := range bsn.hru {
		swprfl := 0. // soil water content of the entire profile excluding the water held in the profile at wilting point [mm] (pg.104)
		for _, ly := range m.sz {
			swprfl += ly.sw - ly.wp
		}
		r += m.surfRunoffLag(m.cn.Update(p, swprfl, false)) // runoff
		inf := math.Max(0., p-r)                            // infiltration
		m.sz[0].sw += inf                                   // add infiltration to soil zone
		i += inf                                            // accumulate subbasin infiltration
		a += m.evap(ep)                                     // actual et
		g += m.percolate()                                  // gw recharge
	}
	b = bsn.baseflow(g) // baseflow
	return
}

// surfRunoffLag returns the the amount of surface runoff released to the main channel (pg.116)
func (m *HRU) surfRunoffLag(qgen float64) float64 {
	qsurf := (qgen + m.qstr) * (1. - math.Exp(-m.surlag/m.tconc))
	m.qstr = qgen + m.qstr - qsurf // update state
	return qsurf
}

func (m *HRU) evap(ep float64) float64 {
	// et := 0.                            // Penman-Monteith is not used; therefore no transpiration
	// ep0 := ep                           // while using SCSCN, canopy extraction is implied (pg.124)
	// es := ep0 * m.cov                   // maximum potential soil evaporation/sublimation (pg.135)
	// eps := math.Min(es, es*ep0/(es+et)) // maximum potential soil evaporation/sublimation adjusted for plant use (pg.136)
	// esub := 0.                          // sublimation has already been accounted for (pg.137)
	// epps := eps - esub                  // maximum potential soil water evaporation adjusted for plant use (pg.137) [mm]
	epps := ep * m.cov // reduced processes from above
	esoilz, esum := make([]float64, nsl+1), 0.
	for i := 0; i <= nsl; i++ {
		z := float64(i) * lythick                              // depth [mm]
		esoilz[i] = epps * z / (z + math.Exp(2.374-0.00713*z)) // evaporative demand at depth z (pg.137)
	}
	for i := 0; i < nsl; i++ {
		esoil := esoilz[i] - esoilz[i+1] // evaporative demand at for layer i (pg.137) // Note ESCO hard-coded to 1.0
		if m.sz[i].sw < m.sz[i].fc {
			esoil *= math.Exp(2.5 * (m.sz[i].sw - m.sz[i].fc) / (m.sz[i].fc - m.sz[i].wp))
		}
		eppsoil := math.Min(esoil, 0.8*(m.sz[i].sw-m.sz[i].wp)) // amount of water removed from layer by evaporation (pg.140) [mm]
		m.sz[i].sw -= eppsoil
		esum += eppsoil
	}
	return esum
}

// percolate from soil zone (pg.151)
func (m *HRU) percolate() float64 {
	w := 0.
	for i, ly := range m.sz {
		if ly.frz {
			w = 0. // soil layer frozen, no percolation allowed
		} else {
			if i < nsl-1 && m.iwt && m.sz[i+1].sw <= m.sz[i+1].fc+(m.sz[i+1].sat-m.sz[i+1].fc)/2. {
				w = 0. // high water table, no percolation allowed
			} else {
				if ly.sw > ly.fc {
					swex := ly.sw - ly.fc                  // [mm]
					w = swex * (1. - math.Exp(-24./ly.tt)) // hard-coded to daily simulations
					m.sz[i].sw -= w
					if i < nsl-1 {
						m.sz[i+1].sw += w
					}
				}
			}
		}
	}
	return w
}

// baseflow is the shallow groundwater aqqounting (deep aquifer not included here)
func (bsn *SubBasin) baseflow(g float64) float64 {
	// note: no bypass flow; no partioning to deep aquifer (pg.173)
	d1 := math.Exp(-1. / bsn.dgw)
	bsn.wrch += (1.-d1)*g + d1*bsn.wrch // recharge entering aquifer (pg.172)
	if bsn.aq > bsn.aqt {
		d1 = math.Exp(-bsn.agw) // only applicable for daily simulations
		bsn.qbf = bsn.qbf*d1 + bsn.wrch*(1.-d1)
	} else {
		bsn.qbf = 0.
	}
	return bsn.qbf
}
