package waterbudget

import "math"

// ABCD monthly water balance model
// see: Alley, W.M., 1984. On the the Treatment of Evapotranspiration, Soil Moisture Accounting, and Aquifer Recharge in Monthly Water Balance Models. Water Resources Research 20(8): 1137-1149.
// based on the abcd model of Thomas (1981)
type ABCD struct {
	a, b, c, d, sto, gwsto float64
}

// New constructor
func (m *ABCD) New(a, b, c, d float64) {
	if a <= 0. || a > 1. || b < 0. || c < 0. || c > 1. || d < 0. {
		panic("ABCD model parameter error")
	}
	m.a = a // the propensity of runoff to occur before the soil is fully saturated
	m.b = b // upper limit on the sum of ET and soil moisture storage
	m.c = c // fraction of surplus water that is set to recharge GW storage ... vs. Alley (1984)'s explanaition: fraction of mean runoff the comes from groundwater
	m.d = d // the reciprocal of the groundwater residence time
}

// Update state
func (m *ABCD) Update(p, ep float64) (float64, float64) {
	w := p + m.sto                                                          // variable W: available water
	y := (w+m.b)/2./m.a - math.Sqrt(math.Pow((w+m.b)/2./m.a, 2.)-w*m.b/m.a) // transfer function (see: TransferFunctions.xlsx)
	// y = a + m.sto // at the end of the month (variable Y); b: upper limit of y

	m.sto = y * math.Exp(-ep/m.b)
	a := y - m.sto
	surp := w - y
	g := m.c * surp // recharge
	m.gwsto = (g + m.gwsto) / (1. + m.d)
	qb := m.d * m.gwsto // groundwater discharge
	q := (1. - m.c) * surp
	return a, q + qb
}
