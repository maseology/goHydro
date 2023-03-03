package rainrun

import (
	"github.com/maseology/goHydro/pet"
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
func (m *MakkinkCCFGR4J) Update(v []float64, doy int) (y, a, r, g float64) {
	const pres = 101300.
	tx, tn, r, s := v[0], v[1], v[2], v[3]

	// calculate yield
	tm := (tx + tn) / 2.
	yt, tf, _ := m.SP.Update(r, s, tm)
	y = yt + tf

	// calculate ep
	ep := func() float64 {
		const (
			a = 0.75
			b = 0.0025
			c = 2.5
		)
		tm := (tx + tn) / 2.
		Kg := m.SI.GlobalFromPotential(tx, tn, a, b, c, doy)
		return pet.Makkink(Kg, tm, pres, m.Palpha, m.Pbeta)
	}()

	a, r, g = m.GR4J.Update(y, ep)
	return
}
