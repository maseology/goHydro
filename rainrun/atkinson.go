package rainrun

import (
	"math"
)

// Atkinson simple storage model, meant for hourly timesteps
// based on formulation given in: Atkinson, S.E., M. Sivapalan, N.R. Viney, R.A. Woods, 2003. Predicting space-time variability of hourly streamflow and the role of climate seasonality: Mahurangi Catchment, New Zealand. Hydrological Processes 17. pp. 2171-2193.
// original ref: Atkinson S.E., R.A. Woods, M. Sivapalan, 2002. Climate and landscape controls on water balance model complexity over changing timescales. Water Resource Research 38(12): 1314.
// additional ref: Wittenberg H., M. Sivapalan, 1999. Watershed groundwater balance equation using streamflow recession analysis and baseflow separation. Journal of Hydrology 219, pp.20-33.
// sto: current storage; sint current interception storage; cov: fractional forest cover; kb = 1/Tcbf
type Atkinson struct {
	sto, sint, sintc, cov, kb, a, b, sbc, sfc float64
}

// New Atkinson constructor
// [sbc, sfc, coverdense, intcap, kb, a, b]
func (m *Atkinson) New(p ...float64) {
	if p[4] < 0. || p[4] > 1. || p[6] < 0. || p[6] > 1. || p[0] < p[1] {
		println(p[0], p[1], p[4], p[6])
		panic("Atkinson input error")
	}
	m.sbc = p[0]           // A.1 - bucket capacity Sbc=D(n-tr)
	m.sfc = p[1]           // A.2 & A.3 - threshold storage; originally written as  Sfc=Sbc*(fc-tr)/(n-tr)=D(fc-tr)
	m.cov = p[2]           // fractional forest density
	m.sintc = p[3]         // interception storage capacity
	m.kb = p[4]            // baseflow recession coefficient
	m.a = p[5]             // sub-surface flow coefficient (S=aQ^b - Wittenberg and Sivapalan, 1999)
	m.b = 1. / (1. - p[6]) // sub-surface flow coefficient [0,1]; reciprocal taken here as opposed to in Update method
}

// Storage returns total storage
func (m *Atkinson) Storage() float64 {
	return m.sto + m.sint
}

// Update state for daily inputs
func (m *Atkinson) Update(p, ep float64) (float64, float64, float64) {
	var a, q, g float64
	ph, eph := p/24., ep/24.
	for i := 0; i < 24; i++ {
		a1, q1, g1 := m.UpdateHourly(ph, eph)
		a += a1
		q += q1
		g += g1
	}
	return a, q, g
	// return m.UpdateHourly(p, ep)
}

// UpdateHourly update state at the intended hourly interval
func (m *Atkinson) UpdateHourly(p, ep float64) (float64, float64, float64) {
	g := m.sto                           // saving antecedent storage
	eveg, ebs := m.cov*ep, (1.-m.cov)*ep // A.4 transpiration; A.5 bare soil evaporation
	if m.sto < m.sfc {
		eveg *= m.sto / m.sfc
	}
	if m.sto < m.sbc {
		ebs *= m.sto / m.sbc
	}
	var qse, qss float64 // A.6 saturation excess, A.7 subsurface runoff
	if m.sto > m.sbc {
		qse = m.sto - m.sbc
		qss = math.Pow((m.sbc-m.sfc)/m.a, m.b)
	} else if m.sto > m.sfc {
		qss = math.Pow((m.sto-m.sfc)/m.a, m.b)
	}
	qbf := m.sto * m.kb // A.8 baseflow
	eint := ep          // A.9 interception evaporation
	if p+m.sint < ep {
		eint = p + m.sint
	}

	a := eveg + ebs + eint // A.10 total actual ET
	var thr float64        // A.13 throughflow
	if p > (m.sintc - m.sint) {
		thr = p - m.sintc + m.sint // modified from original
	}

	m.sint += p - eint - thr // A.11 interception water balance
	if m.sint < mingtzero {
		m.sint = 0.
	}
	m.sto += thr - eveg - ebs - qse - qss - qbf // A.12 soil zone water balance
	if m.sto < mingtzero {
		m.sto = 0.
	}

	g -= m.sto // will produce negative recharge to satisfy ET demand during dry periods, meant for long-term recharge calculations.
	// if g < 0. {
	// 	g = 0. // not part of the Atkinson model, but closest assumption to infiltration (=recharge)
	// }
	q := qse + qss + qbf // total discharge

	return a, q, g
}

// // SampleSpace returns a hypercube from which the optimum resides
// func (m *Atkinson) SampleSpace(u []float64) []float64 {
// 	const sd, n, fc = 1000.0, 0.3, 0.1
// 	x1 := mm.LinearTransform(0., sd*fc, u[1])     // threshold storage (sfc=D(fc-tr))
// 	x0 := x1 + mm.LinearTransform(0., sd*n, u[0]) // watershed storage (sbc=D(n-tr))
// 	x2 := mm.LinearTransform(0., 1., u[2])        // coverdense
// 	x3 := mm.LinearTransform(0., 0.01, u[3])      // intcap
// 	x4 := mm.LinearTransform(0.0001, 1., u[4])    // kb
// 	x5 := mm.LinearTransform(0., 100., u[5])      // a
// 	x6 := mm.LinearTransform(0., 1., u[6])        // b
// 	return []float64{x0, x1, x2, x3, x4, x5, x6}
// }

// // Ndim returns the number of dimensions
// func (m *Atkinson) Ndim() int { return 7 }
