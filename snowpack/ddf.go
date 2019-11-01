package snowpack

// DDF degree-day factor smowmelt method
type DDF struct {
	snowpack
	ddf, ddfc float64
}

// NewDefaultDDF returns a new CCF struct
func NewDefaultDDF() DDF {
	d := DDF{
		ddf:  0.0045, // DDF temperature index; range .001 to .008 m/°C/d  (pg.275)
		ddfc: 1.1,    // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	}
	d.tb = 0.  // base/critical temperature (°C)
	d.tsf = .5 // TSF (surface temperature factor), 0.1-0.5 have been used
	return d
}

// Update state
func (d *DDF) Update(r, s, t float64) (drainage float64) {
	d.addToPack(s, SnowFallDensity(t))
	melt := d.ddf * df * (t - d.tb) // [m·°C-1·d-1]
	if melt < 0. {
		melt = 0.
	}
	d.lwc += melt + r
	d.swe += r
	drainage = d.drainFromPack()
	return
}

func (d *DDF) adjustDegreeDayFactor(den float64) {
	// see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	d.ddf = d.ddfc * den / pw / 100. // [m/(°C·d)]
}
