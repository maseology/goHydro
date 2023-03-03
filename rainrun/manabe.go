package rainrun

// manabe reservoir
// standard form of a hydrological "bucket" model
// ref: Manabe, S., 1969. Climate and the Ocean Circulation 1: The Atmospheric Circulation and The Hydrology of the Earth's Surface. Monthly Weather Review 97(11). 739-744.
type manabe struct {
	res
	expo, minsto float64
}

// new constructor
func (m *manabe) new(capacity, fexposed, minSto float64) {
	if capacity < 0. || minSto < 0. || minSto > capacity || fexposed < 0.0 {
		panic("Manabe parameter error")
	}
	m.cap = capacity
	m.minsto = minSto
	m.expo = fexposed
}

// UpdateExposure : change area of reservoir exposed to evaporative forcings
func (m *manabe) updateExposure(newExposure float64) {
	if newExposure < 0. {
		panic("UpdateExposure error")
	}
	if m.expo > 0. { // for example, in cases of seasonal LAI changes, storage capacity will also change
		fsto := newExposure / m.expo
		m.updateCapacity(fsto)
	}
	m.expo = newExposure
}

// updateCapacity : changes reservoir capacity
// if this causes and reservoir overflow (ie, ChangeFactor < 1), it will be determined with the next state update
func (m *manabe) updateCapacity(changeFactor float64) {
	m.cap *= changeFactor
}

// Update state
func (m *manabe) update(p, ep, perc float64) (float64, float64, float64) {
	q := m.overflow(p)
	if m.sto == 0. {
		return 0., q, 0.
	}
	if ep < mingtzero && perc < mingtzero {
		return 0., q, 0.
	}
	a, g := m.lossDirect(ep, perc)
	// a, g := m.lossExponential(ep, perc, 86400.)
	return a, q, g
}

func (m *manabe) lossDirect(ep, perc float64) (float64, float64) {
	var a, g float64
	epx := m.expo * ep * m.storageFraction() // effective PE
	if m.sto <= m.minsto {
		if ep < mingtzero {
			return 0., 0.
		}
		if epx >= m.sto {
			a = m.sto
			m.sto = 0.
		} else {
			a = epx
			m.sto -= epx
		}
	} else {
		if (epx + perc) > (m.sto - m.minsto) {
			fperc := perc / (epx + perc)
			g = fperc * (m.sto - m.minsto)
			a = m.sto - m.minsto - g
			epx -= a // reset to remaining available PE
			if epx >= m.minsto {
				a += m.minsto
				m.sto = 0.
			} else {
				a += epx
				m.sto = m.minsto - epx
			}
		} else {
			a = epx
			g = perc
			m.sto -= (a + g)
		}
	}
	return a, g
}

func (m *manabe) lossExponential(ep, perc, ts float64) (float64, float64) {
	var a, g float64
	// first compute (direct) drainage
	sFree := m.sto - m.minsto
	if sFree > 0. {
		if sFree <= perc {
			m.sto = m.minsto
			g = sFree
		} else {
			m.sto -= perc
			g = perc
		}
	}
	// next compute AET
	a = m.res.decayExp2(ep/ts, ts)
	return a, g
}

// ManabeGW manabe reserveroir with an added exponential decay reservoir
type ManabeGW struct {
	r              manabe
	gwsto, perc, k float64
}

// New ManabeGW constructor
// [capacity, fexposed, minSto, perc, kbf]
func (m *ManabeGW) New(p ...float64) {
	m.r.new(p[0], p[1], p[2])
	m.perc = p[3]
	m.k = p[4]
}

// Update state for daily inputs
func (m *ManabeGW) Update(p, ep float64) (float64, float64, float64) {
	a, q1, g := m.r.update(p, ep, m.perc)
	m.gwsto += g
	q2 := m.gwsto * (1. - m.k)
	m.gwsto -= q2
	return a, q1 + q2, g
}

// Storage returns manabe storage
func (m *ManabeGW) Storage() float64 {
	return m.r.Storage() + m.gwsto
}

// // SampleSpace returns a hypercube from which the optimum resides
// func (m *ManabeGW) SampleSpace(u []float64) []float64 {
// 	const esdx = 1. // effective maximum soildepth
// 	jn := jointdist.Nested(u[2], u[0])
// 	u2t, u0t := jn[0], jn[1]
// 	x0 := mmaths.LinearTransform(0., esdx, u0t)
// 	x1 := u[1]
// 	x2 := mmaths.LinearTransform(0., esdx, u2t)
// 	x3 := mmaths.LogLinearTransform(1e-10, 1., u[3])
// 	x4 := u[4]
// 	return []float64{x0, x1, x2, x3, x4}
// }

// // Ndim returns the number of dimensions
// func (m *ManabeGW) Ndim() int { return 5 }
