package swat

import "math"

// baseflow is the shallow groundwater accounting (deep aquifer not included here)
func (bsn *SubBasin) baseflow(g float64) float64 {
	// note: no bypass flow; no partioning to deep aquifer (pg.173)
	d1 := math.Exp(-1. / bsn.dgw)
	bsn.psto += g
	bsn.wrch = (1.-d1)*g + d1*bsn.wrch // recharge entering aquifer (pg.172)
	bsn.psto -= bsn.wrch               // water in "percolation" storage
	if bsn.aq > bsn.aqt {
		e1 := math.Exp(-bsn.agw)                // only applicable for daily simulations
		bsn.qbf = bsn.qbf*e1 + bsn.wrch*(1.-e1) // pg.174
	} else {
		bsn.qbf = 0.
	}
	bsn.aq += bsn.wrch - bsn.qbf
	return bsn.qbf // [mm]
}
