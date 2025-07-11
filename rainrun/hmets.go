package rainrun

import (
	"github.com/maseology/goHydro/convolution"
)

// HMETS model
// Martel, J., Demeester, K., Brissette, F., Poulin, A., Arsenault, R., 2017. HMETS - a simple and efficient hydrology model for teaching hydrological modelling, flow forecasting and climate change impacts to civil engineering students. International Journal of Engineering Education 34, 1307–1316.
type HMETS struct {
	lv, lp                       *res
	gsr, gdr                     *convolution.Convolution
	eteff, fimp, cr, cv, cvp, cp float64
}

// New HMETS constructor
// [eteff, fimp, LVcap, LPcap, cr, cv, cvp, cp, sralpha, srbeta, dralpha, drbeta]
func (m *HMETS) New(p ...float64) {
	const gw0 = 250. // mm/yr

	if len(p) == 0 {
		println(" ** Warning: default HMETS parameters being assigned **")
		m.eteff = 1. // Fraction of the potential evapotranspiration
		m.fimp = 0.  // fraction impervious (not in original HMETS, but in Raven)
		m.cr = .3    // Fraction of the water for surface and delayed runoff
		m.cv = .1    // Fraction of the water for hypodermic flow
		m.cvp = .2   // Fraction of the water for groundwater recharge
		m.cp = .1    // Fraction of the water for groundwater flow
		m.gsr = convolution.NewGammaConvolution(.1, .1, 86400.)
		m.gdr = convolution.NewGammaConvolution(.1, .1, 86400.)
		m.lv = &res{cap: 300.}                                  // Maximum level of the vadose zone [mm]
		m.lp = &res{sto: min(gw0/365.24/m.cp, 500.), cap: 500.} // max basin phreatic zone capacity [mm]
	} else {
		m.eteff = p[0]                                          // Fraction of the potential evapotranspiration
		m.fimp = p[1]                                           // fraction impervious/direct flow
		m.lv = &res{cap: p[2]}                                  // max basin vadose zone capacity [mm]
		m.lp = &res{sto: min(gw0/365.24/p[7], p[3]), cap: p[3]} // max basin phreatic zone capacity [mm]
		m.cr = p[4]                                             // Fraction of the water for surface and delayed runoff (i.e., runoff coefficient)
		m.cv = p[5]                                             // Fraction of the water for hypodermic flow
		m.cvp = p[6]                                            // Fraction of the water for groundwater recharge
		m.cp = p[7]                                             // Fraction of the water for groundwater flow
		m.gsr = convolution.NewGammaConvolution(p[8], p[9], 86400.)
		m.gdr = convolution.NewGammaConvolution(p[10], p[11], 86400.)
	}
	if fracCheck(m.cv+m.cvp) || fracCheck(m.cr) {
		panic("HMETS input error")
	}
}

// Update state for daily inputs
func (m *HMETS) Update(pn, ep float64) (float64, float64, float64) {
	a := m.eteff * ep

	lvsat := min(m.lv.storageFraction(), 1.)
	q := m.fimp * pn                         // direct/impervious surface runoff (not in original HMETS paper)
	ht1 := m.cr * lvsat * (1. - m.fimp) * pn // water available for runoff that will be directed to surface runoff

	infil := (1.-m.fimp)*pn - ht1 - a // Infiltration

	ht2 := 0. // Delayed runoff
	if infil >= 0. {
		ht2 = m.cr * lvsat * lvsat * infil // delayed runoff
	} else {
		a += infil
		infil = 0.
	}

	// Vadose zone
	ht3 := m.cv * m.lv.sto // water for the Hypodermic flow component
	g := m.cvp * m.lv.sto  // Groundwater recharge

	ht2 += m.lv.overflow(infil - ht2 - ht3 - g) // delayed runoff (note "-RETt term removed as it's double counted in eq. 9")

	// Phreatic zone
	ht4 := m.cp * m.lp.sto
	ht2 += m.lp.overflow(g - ht4)

	q += m.gsr.Update(ht1) // Surface runoff
	q += m.gdr.Update(ht2) // Delayed runoff
	q += ht4               // Groundwater flow

	//     rates[0]=infil;     //PONDED->SOIL[0]
	//     rates[1]=direct;    //PONDED->SW
	//     rates[2]=runoff;    //PONDED->CONVOL[0]
	//     rates[3]=delayed;   //PONDED->CONVOL[1]
	return a, q, g
}

// Storage returns total storage
func (m *HMETS) Storage() float64 {
	return m.lv.sto + m.lp.sto
}
