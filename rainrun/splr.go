package rainrun

// SPLR : Simple Parallel Linear Reservoir
// Buytaert, W., and K. Beven, 2011. Models as multiple working hypotheses hydrological simulation of tropical alpine. Hydrological Processes 25. pp. 1784–1799.
// 3-reservoir Tank model
type SPLR struct {
	s1, s2, s3  res
	r12, r23, x float64
}

// New SPLR constructor
// [r12, r23, k1, k2, k3, x]
func (m *SPLR) New(p ...float64) {
	m.r12 = p[0]
	m.r23 = p[1]
	m.s1.new(0., p[2])
	m.s2.new(0., p[3])
	m.s3.new(0., p[4])
	m.x = p[5] // ep factor
}

// Update state for daily inputs, returns excess
func (m *SPLR) Update(p, ep float64) (float64, float64, float64) {
	pn, sv := p-ep*m.x, m.Storage()
	m.s1.update(m.r12 * pn)
	m.s2.update((1. - m.r12) * m.r23 * pn)
	g := (1. - m.r12) * (1. - m.r23) * pn
	m.s3.update(g)
	aet := sv - m.Storage() + p
	return aet, m.s1.decayExp() + m.s2.decayExp() + m.s3.decayExp(), g
}

// Storage returns total storage
func (m *SPLR) Storage() float64 {
	return m.s1.sto + m.s2.sto + m.s3.sto
}

/////////////////////////////////////////////////////////////
////////////////////////////////////OLD//////////////////////
/////////////////////////////////////////////////////////////

// type SPLR struct {
// 	s1, s2, s3              float64
// 	r12, r23, k1, k2, k3, x float64
// }

// // New SPLR constructor
// // [r12, r23, k1, k2, k3, x]
// func (m *SPLR) New(p ...float64) {
// 	m.r12 = p[0]
// 	m.r23 = p[1]
// 	m.k1 = p[2]
// 	m.k2 = p[3]
// 	m.k3 = p[4]
// 	m.x = p[5] // ep factor
// }

// // // Update state for daily inputs, returns excess
// // func (m *SPLR) Update(p, ep float64) (float64, float64, float64) {
// // 	if ep*m.x > p {
// // 		ep = p/m.x
// // 	}
// // 	pn := p - ep*m.x
// // 	m.s1 += m.r12 * pn
// // 	m.s2 += (1. - m.r12) * m.r23 * pn
// // 	m.s3 += (1. - m.r12) * (1. - m.r23) * pn
// // 	return ep, q(&m.s1, m.k1) + q(&m.s2, m.k2) + q(&m.s3, m.k3), pn
// // }

// // Update state for daily inputs, returns excess
// func (m *SPLR) Update(p, ep float64) (float64, float64, float64) {
// 	pn, sv := p-ep*m.x, m.Storage()
// 	u(&m.s1, m.r12*pn)
// 	u(&m.s2, (1.-m.r12)*m.r23*pn)
// 	u(&m.s3, (1.-m.r12)*(1.-m.r23)*pn)
// 	aet := sv - m.Storage() + pn
// 	// g := pn * (m.r12 + (1.-m.r12)*m.r23 + (1.-m.r12)*(1.-m.r23))
// 	// if g < 0. {
// 	// 	g = 0.
// 	// }
// 	return aet, q(&m.s1, m.k1) + q(&m.s2, m.k2) + q(&m.s3, m.k3), pn
// }

// func u(s *float64, p float64) {
// 	*s += p
// 	if *s < 0. {
// 		*s = 0.
// 	}
// }

// func q(s *float64, k float64) float64 {
// 	d := k * *s
// 	*s -= d
// 	return d
// }

// // Storage returns total storage
// func (m *SPLR) Storage() float64 {
// 	return m.s1 + m.s2 + m.s3
// }

// // SampleSpace returns a hypercube from which the optimum resides
// func (m *SPLR) SampleSpace(u []float64) []float64 {
// 	r12 := mm.LinearTransform(0., 1., u[0])
// 	r23 := mm.LinearTransform(0., 1., u[1])
// 	k1 := mm.LinearTransform(0., 1., u[2])
// 	k2 := mm.LinearTransform(0., 1., u[3])
// 	k3 := mm.LinearTransform(0., 1., u[4])
// 	return []float64{r12, r23, k1, k2, k3}
// }

// // Ndim returns the number of dimensions
// func (m *SPLR) Ndim() int { return 5 }
