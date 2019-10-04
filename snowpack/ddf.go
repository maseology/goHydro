package snowpack

const (
	ddfc = 1.1 // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
)

// DDF degree-day factor smowmelt method
type DDF struct {
	snowpack
	ddf float64
}

// NewDefaultDDF returns a new CCF struct
func NewDefaultDDF() DDF {
	return DDF{
		ddf: 0.0045, // DDF temperature index; range .001 to .008 m/°C/d  (pg.275)
	}
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
	d.ddf = ddfc * den / pw / 100. // [m/(°C·d)]
}
