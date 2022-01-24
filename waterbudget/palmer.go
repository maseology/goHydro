package waterbudget

import "math"

// Palmer monthly waterbudget model
// see: Alley, W.M., 1984. On the the Treatment of Evapotranspiration, Soil Moisture Accounting, and Aquifer Recharge in Monthly Water Balance Models. Water Resources Research 20(8): 1137-1149.
// The Palmer (1965) model; referred as the P model by Alley (1984), similar to the Thornthwaite and Mather (1955)
type Palmer struct {
	sto, capA, capB, surp, lambda float64
}

// New constructor
func (m *Palmer) New(soilcapLower, soilcapUpper, lag float64) {
	if lag < 0. || lag > 1. || soilcapLower < 0. || soilcapUpper < 0. {
		panic("Thornthwaite Mather parameter error")
	}
	m.capA = soilcapUpper
	m.capB = soilcapLower
	m.lambda = lag // not part of original Palmer method, but has been added here consistent with Alley, 1984; same as the Thornthwaite and Mather (1955)
	// originally, Thornthwaite and Mather (1955) set lambda = 0.5, later Mather (1975) set lambda = 0.75
}

// Update state
func (m *Palmer) Update(p, ep float64) (float64, float64) {
	var a float64
	m.sto += p
	availU := m.sto - m.capB
	if availU > 0. {
		ea := math.Min(availU, ep)
		m.sto -= ea
		a += ea
	}

	if ep-p-a > 0. {
		availL := math.Min(m.sto, m.capB)
		eb := math.Min(availL, m.sto*(ep-p-a)/(m.capA+m.capB))
		m.sto -= eb
		a += eb
	}

	m.surp += math.Max(0., m.sto-m.capA-m.capB)
	q := (1. - m.lambda) * m.surp
	m.surp *= m.lambda
	return a, q
}
