package swat

import "math"

// SCSCN is the soil conservation service (now known as the Natural Resources Conservation Service (NRCS))
// curve number (CN) number runoff generation technique, after the SWAT model implimentation
// defaulted SWAT ICN=0: daily curve number as a function of soil moisture
// ref: Neitsch, S.L., J.G. Arnold, J.R., Kiniry, J.R. Williams, 2011. Soil and Water Assessment Tool: Theoretical Documentation Version 2009 (September 2011). 647pp.
type SCSCN struct {
	cn, smax, w1, w2 float64
}

// New constructor
// cn: cn number; fc: amount of water held at field capacity (mm); sat amount of water held at saturation (mm)
// slp: average fraction slope for the subbasin (CN method assumes a slope fraction of 0.05)
func (c *SCSCN) New(cn, fc, sat, slp float64) {
	ccomp := func(cn float64) (float64, float64) {
		cn1 := cn - (20.*(100.-cn))/(100.-cn+math.Exp(2.533-0.0636*(100.-cn)))
		cn3 := cn * math.Exp(0.00673*(100.-cn))
		return cn1, cn3
	}
	cn1, cn3 := ccomp(cn)
	c.cn = (cn3-cn)*(1.-2.*math.Exp(-13.86*slp))/3. + cn // slope adjustment
	cn1, cn3 = ccomp(c.cn)
	c.smax = 25.4 * (1000/cn1 - 10)
	s3 := 25.4 * (1000/cn3 - 10)
	d := math.Log(fc/(1.-s3/c.smax) - fc)
	c.w2 = (d - math.Log(sat/(1.-2.54/c.smax)-sat)) / (sat - fc) // pg.104
	c.w1 = d + c.w2*fc
}

// Update state. p: precipitation (mm)
// sw: is the soil water content of the entrire profile excluding the amount
// of water held in the profile at wilting point (mm) (pg.104)
func (c *SCSCN) Update(p, sw float64, froz bool) float64 {
	s := c.smax * (1. - sw/(sw+math.Exp(c.w1-c.w2*sw)))
	if froz { // frozen soil adjustment (pg.105)
		s *= (1. - math.Exp(-0.000862*s))
	}
	//cn := 25400. / (s + 254.)
	if p > 0.2*s {
		return math.Pow(p-0.2*s, 2.) / (p + 0.8*s)
	}
	return 0.
}
