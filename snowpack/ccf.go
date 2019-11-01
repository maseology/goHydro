package snowpack

import (
	"log"
	"math"
)

// CCF (cold-content factor) snowpack model
// see pg. 279 in DeWalle, D.R. and A. Rango, 2008. Principles of Snow Hydrology. Cambridge University Press, Cambridge. 410pp.
type CCF struct {
	DDF
	cc, tindex float64
}

// NewDefaultCCF returns a new CCF struct
func NewDefaultCCF() CCF {
	c := CCF{
		tindex: 0.00035, // CCF temperature index; range .0002 to 0.0005 m/°C/d -- roughly 1/10 DDF (pg.278)
	}
	c.ddf = ddfi // degree-day/melt factor; range .001 to .008 m/°C/d  (pg.275) -- NOTE: this is an initial value if adjustDegreeDayFactor() and ddfc is used
	c.ddfc = 1.1 // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	c.tb = 0.    // base/critical temperature (°C)
	c.tsf = .5   // TSF (surface temperature factor), 0.1-0.5 have been used
	return c
}

// NewCCF returns a new CCF struct
func NewCCF(tindex, ddf, ddfc, baseT, tsf float64) CCF {
	c := CCF{
		tindex: tindex, // CCF temperature index; range .0002 to 0.0005 m/°C/d -- roughly 1/10 DDF (pg.278)
	}
	c.ddf = ddf   // degree-day/melt factor; range .001 to .008 m/°C/d  (pg.275) -- NOTE: this is an initial value if adjustDegreeDayFactor() and ddfc is used
	c.ddfc = ddfc // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	c.tb = baseT  // base/critical temperature (°C)
	c.tsf = tsf   // TSF (surface temperature factor), 0.1-0.5 have been used
	return c
}

// Update state
func (c *CCF) Update(r, s, t float64) (drainage float64) {
	blNewPack := c.swe == 0.
	if blNewPack {
		c.ddf = ddfi // re-initialize ddf
	}
	if s > 0. {
		c.addToPack(s, SnowFallDensity(t))
	}
	if blNewPack || s > 0.005 {
		c.ts = math.Min(t, 0.)
	} else {
		c.updateSurfaceTemperature(t)
	}

	c.adjustDegreeDayFactor(c.den)
	melt := c.ddf * df * (t - c.tb) // [m·°C-1·d-1]
	if melt > 0. {
		if melt > c.swe-c.lwc {
			melt = c.swe - c.lwc
			c.internalFreeze(-melt)
			c.lwc = c.swe
			c.cc = 0.
			c.ts = 0.
		} else {
			c.internalFreeze(-melt)
			c.lwc += melt
		}
	} else {
		melt = 0.
	}

	if r > 0. {
		c.addToPack(r, pw)
		c.lwc += r
	}

	c.satisfyColdContent(t)
	drainage = c.drainFromPack()
	if c.swe == 0. {
		c.cc = 0.
	}
	// c.densify() // currently disabled, need to lookup the coefficient to the densification factor
	return
}

func (c *CCF) satisfyColdContent(t float64) {
	if (c.swe - c.lwc) <= 0. {
		if c.swe != c.lwc {
			log.Fatalf("CCF.satisfyColdContent error: swe and lwc should be equivalent.\n  swe = %f;  lwc = %f", c.swe, c.lwc)
		}
		c.swe = c.lwc
		c.cc = 0.
		c.ts = 0.
		if c.swe > 0. {
			c.den = pw
		} else {
			c.den = 0.
		}
	} else {
		c.cc += c.tindex * df * (c.ts - t)
		if c.cc <= 0. {
			c.cc = 0.
			c.ts = 0.
		}
		if c.lwc > 0. && c.cc > 0. {
			if c.lwc >= c.cc { // liquid water available to lower cold content; check to see if pack becomes isothermal
				c.internalFreeze(c.cc)
				c.lwc -= c.cc
				c.cc = 0.
			} else { // all liquid water freezes
				c.internalFreeze(c.lwc)
				c.cc -= c.lwc
				c.lwc = 0.
			}
		}
	}
}

// Properties returns the snowpack state
func (c *CCF) Properties() (porosity, depth float64) {
	return c.properties()
}
