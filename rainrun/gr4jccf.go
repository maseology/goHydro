package rainrun

import (
	"github.com/maseology/goHydro/snowpack"
	"github.com/maseology/goHydro/solirrad"
)

// CCFGR4J model
// Perrin C., C. Michel, V. Andreassian, 2003. Improvement of a parsimonious model for streamflow simulation. Journal of Hydrology 279. pp. 275-289.
type CCFGR4J struct {
	GR4J
	SP snowpack.CCF
	SI *solirrad.SolIrad
}

// New CCFGR4J contructor
// [stocap, gwstocap, x4, unitHydrographPartition, x2]
// [tindex, ddfc, baseT, tsf]
func (m *CCFGR4J) New(p ...float64) {
	const ddf = 0.0045
	// GR4J
	m.GR4J.New(p...)

	// Cold-content snow melt funciton
	tindex, ddfc, baseT, tsf := p[4], p[5], p[6], p[7]
	m.SP = snowpack.NewCCF(tindex, ddf, ddfc, baseT, tsf)
}

// Update state for daily inputs
func (m *CCFGR4J) Update(v *Dset) (y, a, r, g float64) {

	// calculate yield
	tm := (v.Tx + v.Tn) / 2.
	yt, tf, _ := m.SP.Update(v.rf, v.sf, tm)
	y = yt + tf

	a, r, g = m.GR4J.Update(y, v.Ep)
	return
}
