package rainrun

import (
	"github.com/maseology/goHydro/snowpack"
	"github.com/maseology/goHydro/solirrad"
	"github.com/maseology/goHydro/transfunc"
)

// CCFHBV model
// Bergström, S., 1976. Development and application of a conceptual runoff model for Scandinavian catchments. SMHI RHO 7. Norrköping. 134 pp.
// Bergström, S., 1992. The HBV model - its structure and applications. SMHI RH No 4. Norrköping. 35 pp
type CCFHBV struct {
	HBV
	SP snowpack.CCF
	SI *solirrad.SolIrad
}

// New CCFHBV constructor
// [fc, lp, beta, uzl, k0, k1, k2, ksat, maxbas, lakeCoverFrac, tindex, ddfc, baseT, tsf]
func (m *CCFHBV) New(p ...float64) {
	const ddf = 0.0045
	if fracCheck(p[1]) || fracCheck(p[4]) || fracCheck(p[5]) || fracCheck(p[6]) { // || fracCheck(p[9]) {
		panic("HBV input eror")
	}
	m.fc = p[0]                         // max basin moisture storage
	m.lp = p[1]                         // soil moisture parameter
	m.beta = p[2]                       // soil moisture parameter
	m.uzl = p[3]                        // upper zone fast flow limit
	m.k0, m.k1, m.k2 = p[4], p[5], p[6] // fast, slow, and baseflow recession coefficients
	m.perc = p[7]                       // upper-to-lower zone percolation, assuming percolation rate = Ksat
	m.lakefrac = 0.                     //p[9]                   // lake fraction

	m.tf = transfunc.NewTF(p[8], 0.5, 0.) // MAXBAS: triangular weighted transfer function

	// Cold-content snow melt funciton
	tindex, ddfc, baseT, tsf := p[9], p[10], p[11], p[12]
	m.SP = snowpack.NewCCF(tindex, ddf, ddfc, baseT, tsf)
}

// Update state
func (m *CCFHBV) Update(v *Dset) (y, a, r, g float64) {

	// calculate yield
	tm := (v.Tx + v.Tn) / 2.
	yt, tf, _ := m.SP.Update(v.rf, v.sf, tm)
	y = yt + tf

	a, r, g = m.HBV.Update(y, v.Ep)
	// a = ep
	// r = y
	return
}
