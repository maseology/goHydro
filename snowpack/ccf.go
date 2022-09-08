package snowpack

import (
	"log"
	"math"
)

// CCF (cold-content factor) snowpack model
// see pg. 279 in DeWalle, D.R. and A. Rango, 2008. Principles of Snow Hydrology. Cambridge University Press, Cambridge. 410pp.
type CCF struct {
	DDF
	cc, ccf, ts, tsf float64
}

// NewDefaultCCF returns a new CCF struct
func NewDefaultCCF() CCF {
	c := CCF{
		ccf: 0.00035, // CCF temperature index; range .0002 to 0.0005 m/°C/d -- roughly 1/10 DDF (pg.278)
		tsf: .5,      // TSF (surface temperature factor), 0.1-0.5 have been used
	}
	c.ddf = ddfi // degree-day/melt factor; range .001 to .008 m/°C/d  (pg.275) -- NOTE: this is an initial value if adjustDegreeDayFactor() and ddfc is used
	c.ddfc = 1.1 // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	c.tb = 0.    // base/critical temperature (°C)
	return c
}

// NewCCF returns a new CCF struct
func NewCCF(ccf, ddfc, baseT, tsf, denscoef float64) CCF {
	c := CCF{
		ccf: ccf, // CCF temperature index; range .0002 to 0.0005 m/°C/d -- roughly 1/10 DDF (pg.278)
	}
	c.ddf = ddfi          // degree-day/melt factor; range .001 to .008 m/°C/d  (pg.275) -- NOTE: this is an initial value if adjustDegreeDayFactor() and ddfc is used
	c.ddfc = ddfc         // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	c.tb = baseT          // base/critical temperature (°C)
	c.tsf = tsf           // TSF (surface temperature factor), 0.1-0.5 have been used
	c.denscoef = denscoef // coefficient to the densification factor
	return c
}

// Update state
func (c *CCF) Update(r, s, t float64) (melt, throughfall float64, err error) {
	err = nil
	if err = inputDataCheck(r, s, t); err != nil {
		return
	}
	// fmt.Println(c.Properties())
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

	c.adjustDegreeDayFactor()
	potmelt := c.ddf * df * (t - c.tb) // [m·°C-1·d-1]
	if potmelt > 0. {
		if potmelt >= c.swe-c.lwc {
			potmelt = c.swe - c.lwc
			c.internalFreeze(-potmelt)
			c.lwc = c.swe
			c.cc = 0.
			c.ts = 0.
		} else {
			c.internalFreeze(-potmelt)
			c.lwc += potmelt
		}
	} else {
		potmelt = 0.
	}

	if r > 0. {
		c.addToPack(r, pw)
		c.lwc += r
	}

	drn := c.drainFromPack()
	c.satisfyColdContent(t)
	c.densify()

	if c.swe <= 0. {
		c.cc = 0.
		c.swe = 0.
	}

	if drn > r {
		throughfall = r
		melt = drn - r
	} else if drn < r {
		throughfall = drn
		melt = 0.
	} else {
		throughfall = r
		melt = 0.
	}
	return
}

func (c *CCF) updateSurfaceTemperature(t float64) { // pg.279
	if c.swe > 0. {
		c.ts += c.tsf * df * (t - c.ts)
		if c.ts > 0. {
			c.ts = 0.
		}
	} else {
		c.ts = 0.
	}
}

func (c *CCF) satisfyColdContent(t float64) {
	if (c.swe - c.lwc) <= 0. { // excess liquid water
		if math.Abs((c.swe-c.lwc)/c.lwc) > 1.e-8 {
			log.Fatalf("CCF.satisfyColdContent error: swe and lwc should be equivalent:  swe = %f;  lwc = %f", c.swe, c.lwc)
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
		c.cc += c.ccf * df * (c.ts - t)
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
func (c *CCF) Properties() (porosity, depth, swe, cc float64) {
	porosity, depth = c.properties()
	swe = c.swe
	cc = c.cc
	return
}
