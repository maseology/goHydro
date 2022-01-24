package waterbudget

import "math"

// NOPEX6 monthly water and snow balance model
// ref: Xu, C.-Y., V.P. Singh, 2005. Evaluation of three complementary relationship evapotranspiration models by water balance approach to estimate actual regional evapotranspiration in different climate regions. Journal of Hydrology 308. pp.105-121.
type NOPEX6 struct {
	sm, sp, a1, a2, a3, a4, a5, a6 float64
}

// New constructor
func (m *NOPEX6) New(a1, a2, a3, a4, a5, a6 float64) {
	if a1 < a2 || a4 < 0. || a4 > 1. || a5 < 0. || a6 < 0. {
		panic("NOPEX6 input error")
	}
	m.a1 = a1
	m.a2 = a2
	m.a3 = a3
	m.a4 = a4
	m.a5 = a5
	m.a6 = a6
}

//'Pmon, PETmon, Tmon, and TmonLT are monthly precipitation, long term monthly PET, temperature, and long term monthly temperature
//'SPt: snow pack storage;

// Update state
func (m *NOPEX6) Update(p, epLongterm, t, tLongterm float64) (float64, float64) {

	// snow fall and rainfall [a1,a2]
	snowtot := math.Max(1.-math.Pow(math.Exp(-(t-m.a1)/(m.a1-m.a2)), 2.), 0.) * p // snow added
	raintot := p - snowtot                                                        // rain added
	m.sp += snowtot                                                               // snowpack storage

	// snow melt [a1,a2]
	mt := math.Max(1.-math.Pow(math.Exp((t-m.a2)/(m.a1-m.a2)), 2.), 0.) * m.sp //snowpack melt
	m.sp -= mt

	// ET [a3, a4]
	ept := (1. + m.a3*(t-tLongterm)) * epLongterm        // potential ET
	wt := raintot + m.sm                                 // available water
	a := math.Min(ept*(1.-math.Pow(m.a4, (wt/ept))), wt) // AET

	// slowflow (GW recharge) [a5]
	qb := m.a5 * m.sm // error in Xu and Singh (2005)

	// fast flow [a6]
	nt := raintot - ept*(1.-math.Exp(-raintot/ept)) // active rainfall
	qf := m.a6 * m.sm * (mt + nt)                   // error in Xu and Singh (2005)

	// total runoff
	q := qb + qf

	// update soil moisture
	m.sm += raintot + mt - a - q
	if m.sm < 0. {
		m.sm = 0.
	}

	return a, q
}
