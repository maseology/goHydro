package waterbudget

import "math"

// ThornthwaiteMather monthly waterbudget model
// see: Alley, W.M., 1984. Onthe the Treatment of Evapotranspiration, Soil Moisture Accounting, and Aquifer Recharge in Monthly Water Balance Models. Water Resources Research 20(8): 1137-1149.
// variant of the Thornthwaite and Mather (1955) model.  Referred as the type T model by Alley (1984)
type ThornthwaiteMather struct {
	sto, cap, surp, lambda float64
}

// New constructor
func (m *ThornthwaiteMather) New(soilcap, lag float64) {
	if lag < 0. || lag > 1. || soilcap < 0. {
		panic("Thornthwaite Mather parameter error")
	}
	m.cap = soilcap
	m.lambda = lag // Time Lag Fraction, originally, Thornthwaite and Mather (1955) set lambda = 0.5, later Mather (1975) set lambda = 0.75
}

// Update state
func (m *ThornthwaiteMather) Update(p, ep float64) (float64, float64) {
	var a float64
	if p >= ep {
		avail := p - ep + m.sto
		m.sto = math.Min(avail, m.cap)
		a = ep
		m.surp += avail - m.sto
	} else {
		// here, the original Thornthwaite and Mather (1955) model makes the use of tables and graphs (Alley, 1984), Alley (1984) offers an analytical solution
		avail := p + m.sto
		m.sto *= math.Exp(-(ep - p) / m.cap)
		a = avail - m.sto
	}
	q := (1. - m.lambda) * m.surp
	m.surp *= m.lambda
	return a, q
}
