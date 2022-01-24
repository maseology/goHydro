package waterbudget

// ABC : annual waterbudget model
// p.69 in: Fiering, 1967. Streamflow Synthesis. Macmillan and Company Ltd., London.
// Recharge = aPt; ET loss = bPt; RO = (1-a-b)Pt; Baseflow = cS(t-1)
// NOTE: originally meant for use on annual timescales.
type ABC struct {
	a, b, c, sto float64
}

// New : constructor
func (m *ABC) New(a, b, c float64) {
	if a+b > 1. {
		panic("ABC model parameter error, a+b>1.0")
	}
	m.a = a
	m.b = b
	m.c = c
}

// Update state
func (m *ABC) Update(p, ep float64) (float64, float64) {
	// g := m.a * p // recharge
	a := m.b * p               // aet
	q := (1.0 - m.a - m.b) * p // overland flow
	qb := m.c * m.sto          // baseflow
	m.sto = (1.0-m.c)*m.sto + m.a*p
	return a, q + qb
}
