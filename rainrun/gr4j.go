package rainrun

import (
	"log"
	"math"

	"github.com/maseology/glbopt"
	"github.com/maseology/mmaths"
)

// GR4J model
// Perrin C., C. Michel, V. Andreassian, 2003. Improvement of a parsimonious model for streamflow simulation. Journal of Hydrology 279. pp. 275-289.
type GR4J struct {
	prd, rte           res
	uh1, uh2, cv1, cv2 []float64
	x2, qsplt          float64
}

// New GR4J constructor
func (m *GR4J) New(p ...float64) {
	if p[3] < .5 { //|| p[4] <= 0. || p[4] >= 1. {
		log.Fatalln("GR4J input error")
	}

	m.prd.new(p[0], 0.) // prd: x1: maximum capacity of the "production (SMA) store"
	m.x2 = p[1]         // x2: water exchange coefficient (>0 for water imports, <0 for exports, =0 for no exchange)
	m.rte.new(p[2], 0.) // rte: x3: reference capacity of "routing store"
	x4 := p[3]          // x4: unit hydrograph time parameter
	// m.qsplt = p[4]      // qsplt: unitHydrographPartition, fixed in paper to = 0.9

	m.rte.sto = func() float64 {
		q0 := p[len(p)-1] // frc.Dat[0][2]
		smpl := func(u float64) float64 {
			return mmaths.LinearTransform(0., 10., u)
		}
		opt := func(u float64) float64 {
			x3i := smpl(u)
			qr := p[1] * math.Pow(x3i/p[2], 7./2.)                       // eq.18 catchment GW exchange; x2: water exchange coefficient (>0 for water imports, <0 for exports, =0 for no exchange)
			qr += x3i * (1. - math.Pow(1.+math.Pow(x3i/p[2], 4.), -.25)) // eq.20
			return math.Abs(qr-q0) / q0
		}
		u, _ := glbopt.Fibonacci(opt)
		return smpl(u)
	}()

	// unit hydrographs build
	func() { // build UH1
		n := int(math.Ceil(x4))
		m.uh1 = make([]float64, n)
		m.cv1 = make([]float64, n-1)
		sh1 := make([]float64, n)
		for t := 1; t < n; t++ {
			tf := float64(t)
			if tf < x4 {
				sh1[t] = math.Pow(tf/x4, 2.5)
			} else {
				sh1[t] = 1.
			}
		}
		for t := 0; t < n; t++ {
			if t < n-1 {
				m.uh1[t] = sh1[t+1] - sh1[t]
			} else {
				m.uh1[t] = 1. - sh1[t]
			}
		}
	}()
	func() { // build UH2
		n := int(math.Ceil(2. * x4))
		m.uh2 = make([]float64, n)
		m.cv2 = make([]float64, n-1)
		sh2 := make([]float64, n)
		for t := 1; t < n; t++ {
			tf := float64(t)
			if tf <= x4 {
				sh2[t] = math.Pow(tf/x4, 2.5) / 2.
			} else if tf < 2.*x4 {
				sh2[t] = 1. - math.Pow(2.-tf/x4, 2.5)/2.
			} else {
				sh2[t] = 1.
			}
		}
		for t := 0; t < n; t++ {
			if t < n-1 {
				m.uh2[t] = sh2[t+1] - sh2[t]
			} else {
				m.uh2[t] = 1. - sh2[t]
			}
		}
	}()
}

// Update state for daily inputs
func (m *GR4J) Update(p, ep float64) (float64, float64, float64) {
	var pn, en, es float64
	if p >= ep {
		pn = p - ep // eq.1
	} else {
		en = ep - p // eq.2 available PET
	}
	x1 := m.prd.cap // x1: maximum capacity of the SMA store
	d1 := math.Tanh(pn / x1)
	sf := m.prd.storageFraction()
	ps := x1 * d1 * (1. - math.Pow(sf, 2.)) / (1. + d1*sf) // eq.3 Ps: portion of rain infiltrating soils (production) store
	if en > 0. {
		d1 = math.Tanh(en / x1)
		es = m.prd.sto * d1 * (2. - sf) / (1. + d1*(1.-sf)) // eq.4 Es: soil evaporation
	}
	m.prd.update(ps - es) // eq.5
	if m.prd.storageFraction() > 1.000001 {
		println(m.prd.sto, m.prd.cap, ps, es)
		panic("GR4J error: production store error")
	}

	g := m.prd.sto * (1. - math.Pow(1.+math.Pow(4.*m.prd.storageFraction()/9., 4.), -0.25)) // eq.6 "Perc": percolation from production zone
	if m.prd.update(-g) < 0. {                                                              // eq.7 this line must be left here such that prd is updated
		panic("GR4J error: percolation")
	}

	pr := g + pn - ps          // eq.8
	q9 := m.updateUH1(.9 * pr) // eq.9-11
	q1 := m.updateUH2(.1 * pr) // eq.12-17
	// q9 := m.updateUH1(m.qsplt * pr)        // eq.9-11
	// q1 := m.updateUH2((1. - m.qsplt) * pr) // eq.12-17

	fe := m.x2 * math.Pow(m.rte.storageFraction(), 7./2.)                              // eq.18 catchment GW exchange; x2: water exchange coefficient (>0 for water imports, <0 for exports, =0 for no exchange)
	m.rte.update(q9 + fe)                                                              // eq.19
	qr := m.rte.sto * (1. - math.Pow(1.+math.Pow(m.rte.storageFraction(), 4.), -0.25)) // eq.20
	if m.rte.update(-qr) < 0. {                                                        // eq.21 this line must be left here such that rte is updated
		panic("GR4J error: routing")
	}

	qd := math.Max(0., q1+fe) // eq.22
	return es, qd + qr, g     // eq.23
}

func (m *GR4J) updateUH1(pr float64) float64 {
	n := len(m.cv1) - 1
	if n == -1 {
		return pr
	}
	q := m.uh1[0]*pr + m.cv1[0]
	for i := 0; i < n; i++ {
		m.cv1[i] = m.uh1[i+1]*pr + m.cv1[i+1]
	}
	m.cv1[n] = m.uh1[n+1] * pr
	return q
}
func (m *GR4J) updateUH2(pr float64) float64 {
	q := m.uh2[0]*pr + m.cv2[0]
	n := len(m.cv2) - 1
	for i := 0; i < n; i++ {
		m.cv2[i] = m.uh2[i+1]*pr + m.cv2[i+1]
	}
	m.cv2[n] = m.uh2[n+1] * pr
	return q
}

// Storage returns total storage
func (m *GR4J) Storage() float64 {
	return m.prd.sto + m.rte.sto
}
