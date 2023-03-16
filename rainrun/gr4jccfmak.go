package rainrun

import (
	"github.com/maseology/goHydro/snowpack"
	"github.com/maseology/goHydro/solirrad"
)

// MakkinkCCFGR4J model
// Perrin C., C. Michel, V. Andreassian, 2003. Improvement of a parsimonious model for streamflow simulation. Journal of Hydrology 279. pp. 275-289.
// with CCF snowmelt model and Makkink PET
type MakkinkCCFGR4J struct {
	GR4J
	SP            snowpack.CCF
	SI            *solirrad.SolIrad
	Palpha, Pbeta float64
}

// New CCFGR4J contructor
// [stocap, gwstocap, x4, unitHydrographPartition, x2]
// [tindex, ddfc, baseT, tsf]
// [b, c, alpha, beta]
func (m *MakkinkCCFGR4J) New(p ...float64) {
	const ddf = 0.0045
	// GR4J
	m.GR4J.New(p...)

	// Cold-content snow melt funciton
	tindex, ddfc, baseT, tsf := p[4], p[5], p[6], p[7]
	m.SP = snowpack.NewCCF(tindex, ddf, ddfc, baseT, tsf)
	m.Palpha, m.Pbeta = p[8], p[9]
}

// Update state for daily inputs
func (m *MakkinkCCFGR4J) Update(d *Dset) (y, a, r, g float64) {
	const pres = 101300.

	// calculate yield
	tm := (d.Tx + d.Tn) / 2.
	yt, tf, _ := m.SP.Update(d.rf, d.sf, tm)
	y = yt + tf

	a, r, g = m.GR4J.Update(y, d.Ep)
	return
}
