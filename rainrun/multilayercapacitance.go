package rainrun

import (
	"math"
)

// MultiLayerCapacitance model
// ref: Struthers, I., C. Hinz, M. Sivapalan, G. Deutschmann, F. Beese, R. Meissner, 2003. Modelling the water balance of a free-draining lysimeter using the downward approach. Hydrological Processes (17). pp. 2151-2169.
// modification here: _runoff = lateral flow (runoff & subsurface)
type MultiLayerCapacitance struct {
	s1, s2, s3, bf                      res
	a1, a2, a3, binv, cv, fc1, fc2, fc3 float64
}

// New MultiLayerCapacitance constructor
// [coverDens, szDepth, porosity, fc, a, b, l1, l2, l3]
func (m *MultiLayerCapacitance) New(p ...float64) {
	if len(p) == 0 {
		println(" ** Warning: default MultiLayerCapacitance parameters being assigned **")
		p = make([]float64, 10)
		p[0] = .1   // coverDens
		p[1] = 350. // szDepth
		p[2] = .3   // porosity
		p[3] = .01  // fc
		p[4] = 90   // a
		p[5] = .6   // b
		p[6] = 1.   // l1
		p[7] = 0.   // l2
		p[8] = 0.   // l3
		p[9] = .95  // baseflow recession
	}
	if math.Abs(p[6]+p[7]+p[8]-1) > 1e-6 || fracCheck(p[0]) || fracCheck(p[2]) || fracCheck(p[3]) || fracCheck(p[9]) { //|| p[3] > p[2] {
		panic("MultiLayerCapacitance input error")
	}
	m.cv = p[0]         // fraction vegetation cover
	fc := p[3] / p[2]   // fraction tension storage
	smax := p[1] * p[2] // total soil zone storage
	m.s1.new(p[6]*smax, 0.)
	m.s2.new(p[7]*smax, 0.)
	m.s3.new(p[8]*smax, 0.)
	m.fc1 = p[6] * smax * fc
	m.fc2 = p[7] * smax * fc
	m.fc3 = p[8] * smax * fc
	m.a1 = p[6] * p[4]
	m.a2 = p[7] * p[4]
	m.a3 = p[8] * p[4]
	m.binv = 1. / p[5]
	m.bf.new(0., 1.-p[9])
}

// Update state for daily inputs
func (m *MultiLayerCapacitance) Update(p, ep float64) (float64, float64, float64) {
	a, q, g := 0., 0., 0.

	g1 := 0.
	if m.a1 > 0 {
		// layer 1
		if m.s1.sto > m.fc1 {
			g1 = math.Pow(((m.s1.sto - m.fc1) / m.a1), m.binv) // drainage from layer 1 to 2
		}
		ebare := (1. - m.cv) * ep * m.s1.storageFraction()                 // bare soil evaporation
		etransp := m.cv * ep * math.Min(m.s1.sto, m.fc1) / (m.fc1 + m.fc2) // transpiration (shallow)
		q += m.s1.overflow(p - ebare - etransp - g1)                       // runoff
		a += ebare + etransp
		g = g1
	}

	g2 := 0.
	if m.a2 > 0 {
		// layer 2
		if m.s2.sto > m.fc2 {
			g2 = math.Pow(((m.s2.sto - m.fc2) / m.a2), m.binv) // drainage from layer 2 to 3
		}
		etransp := m.cv * ep * math.Min(m.s2.sto, m.fc2) / (m.fc1 + m.fc2) // transpiration (deep)
		q += m.s2.overflow(g1 - etransp - g2)                              // interflow
		a += etransp
		g = g2
	}

	g3 := 0.
	if m.a3 > 0 {
		// layer 3
		if m.s3.sto > m.fc3 {
			g3 = math.Pow(((m.s3.sto - m.fc3) / m.a3), m.binv) // drainage from layer 3 to gw reservoir
		}
		q += m.s3.overflow(g2 - g3) // baseflow
		g = g3
	}

	m.bf.update(g)
	q += m.bf.decayExp() // baseflow
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
