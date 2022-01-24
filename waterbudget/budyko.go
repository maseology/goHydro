package waterbudget

import "math"

// Budyko: monthly waterbudget
// pg 96 in Budyko 1975. Climate and Life
// referenced in Budyko (1975) as Budyko (1950)
// referenced in Manabe (1969) as Budyko (1956)
type Budyko struct {
	sto, cap, mu float64
}

// New : constrcutor
func (m *Budyko) New(cap, mu float64) {
	m.cap = cap
	m.mu = mu // dimensionless proportionality constant, increases with storm intensity (0.2 north of 45°N)
}

// Update : update state
func (m *Budyko) Update(p, ep float64) (float64, float64) {
	var a, q float64
	if p > ep {
		if m.sto == m.cap {
			a = ep
			q = p - ep
		} else {
			eta2 := math.Pow(1.-ep/p, 2.)
			a = ep * m.sto / m.cap
			q = p * m.sto * math.Sqrt(math.Pow((1.-eta2)*m.mu, 2.)+eta2) / m.cap
		}
	} else {
		if m.sto == m.cap {
			a = ep
			m.sto += p - ep
		} else {
			q = m.mu * p * m.sto / m.cap
			a = ep * m.sto / m.cap
		}
	}
	m.sto += p - a - q
	if m.sto > m.cap {
		q += m.sto - m.cap
		m.sto = m.cap
	}
	return a, q
}
