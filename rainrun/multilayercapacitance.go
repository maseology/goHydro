package rainrun

import (
	"math"
)

// MultiLayerCapacitance model
// ref: Struthers, I., C. Hinz, M. Sivapalan, G. Deutschmann, F. Beese, R. Meissner, 2003. Modelling the water balance of a free-draining lysimeter using the downward approach. Hydrological Processes (17). pp. 2151-2169.
// modification here: _runoff = lateral flow (runoff & subsurface)
type MultiLayerCapacitance struct {
	s1, s2, s3            res
	a1, a2, a3, b, cv, fc float64
}

// New MultiLayerCapacitance constructor
// [coverDens, szDepth, porosity, fc, a, b, l1, l2, l3]
func (m *MultiLayerCapacitance) New(p ...float64) {
	if p[6]+p[7]+p[8] != 1. || fracCheck(p[0]) || p[3] < 0. || p[3] > p[2] || p[2] > 0. {
		panic("MultiLayerCapacitance input error")
	}
	m.cv = p[0]         // fraction vegetation cover
	m.fc = p[3] / p[2]  // fraction tension storage
	smax := p[1] * p[2] // total soil zone storage
	m.s1.new(p[6]*smax, 0.)
	m.s2.new(p[7]*smax, 0.)
	m.s3.new(p[8]*smax, 0.)
	m.a1 = p[6] * p[4]
	m.a2 = p[7] * p[4]
	m.a3 = p[8] * p[4]
	m.b = 1. / p[5]
}

// Update state for daily inputs
func (m *MultiLayerCapacitance) Update(p, ep float64) (float64, float64, float64) {
	var q float64
	// layer 1
	g := math.Pow(((m.s1.sto - m.s1.cap*m.fc) / m.a1), m.b)
	e1 := (1.-m.cv)*ep*m.s1.storageFraction() - m.cv*ep*math.Min(m.s1.sto, m.s1.cap*m.fc)
	s1n := m.s1.sto + p - e1/(m.s1.cap+m.s2.cap)/m.fc - g
	if s1n > m.s1.cap {
		q = s1n - m.s1.cap
		s1n = m.s1.cap
	}

	// layer 2
	var s2n, e2 float64
	if m.s2.cap > 0. {
		g = math.Pow(((m.s2.sto - m.s2.cap*m.fc) / m.a2), m.b)
		e2 = m.cv * ep * math.Min(m.s2.sto, m.s2.cap*m.fc)
		s2n = m.s2.sto - e2/(m.s1.cap+m.s2.cap)/m.fc + math.Pow(((m.s1.sto-m.s1.cap*m.fc)/m.a1), m.b) - g
		if s2n > m.s2.cap {
			q += s2n - m.s2.cap
			s2n = m.s2.cap
		}
	}

	// layer 3
	var s3n float64
	if m.s3.cap > 0. {
		g = math.Pow(((m.s3.sto - m.s3.cap*m.fc) / m.a3), m.b)
		s3n = m.s3.sto + math.Pow(((m.s2.sto-m.s2.cap*m.fc)/m.a2), m.b) - g
		if s3n > m.s3.cap {
			q += s3n - m.s3.cap
			s3n = m.s3.cap
		}
	}
	a := e1 + e2
	m.s1.sto = s1n
	m.s2.sto = s2n
	m.s3.sto = s3n
	return a, q, g
}

// Storage returns total storage
func (m *MultiLayerCapacitance) Storage() float64 {
	return m.s1.sto + m.s2.sto + m.s3.sto
}

// // SampleSpace returns a hypercube from which the optimum resides
// func (m *MultiLayerCapacitance) SampleSpace(u []float64) []float64 {
// 	const sd, n = 1000.0, 0.3
// 	cv := mm.LinearTransform(0., 1., u[0])
// 	x1 := mm.LinearTransform(0., sd, u[1])
// 	uj0, uj1 := jointdist.Nested2(u[2], u[3])
// 	x2 := mm.LinearTransform(0., n, uj0)
// 	fc := mm.LinearTransform(0., n, uj1)
// 	a := mm.LinearTransform(0., 100., u[4])
// 	b := mm.LinearTransform(0., 1., u[5])
// 	l := jointdist.SumToOne(u[6], u[7], u[8])
// 	return []float64{cv, x1, x2, fc, a, b, l[0], l[1], l[2]}
// }

// // Ndim returns the number of dimensions
// func (m *MultiLayerCapacitance) Ndim() int { return 9 }
