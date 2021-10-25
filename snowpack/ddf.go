package snowpack

// DDF degree-day factor smowmelt method
type DDF struct {
	snowpack
	ddf, ddfc float64
}

const (
	ddfi = 0.0045
	ddrf = .05 // re-freeze factor Seibert (2005)
)

func (d *DDF) adjustDegreeDayFactor(den float64) {
	// see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	d.ddf = d.ddfc * den / pw // [m/(°C·d)]
}

func NewDDF(ddfc, baseT float64) DDF {
	d := DDF{
		ddf:  ddfi, // degree-day/melt factor; range .001 to .008 m/°C/d  (pg.275)
		ddfc: ddfc, // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
	}
	d.tb = baseT // base/critical temperature (°C)
	return d
}

// // NewDefaultDDF returns a new CCF struct
// func NewDefaultDDF() DDF {
// 	d := DDF{
// 		ddf:  ddfi, // degree-day/melt factor; range .001 to .008 m/°C/d  (pg.275)
// 		ddfc: 1.1,  // DDF adjustment factor based on pack density, see DeWalle and Rango, pg. 275; Ref: Martinec (1960)
// 	}
// 	d.tb = 0. // base/critical temperature (°C)
// 	return d
// }

// Update state
func (d *DDF) Update(r, s, t float64) (drainage float64) {
	inputDataCheck(r, s, t)

	if s > 0. {
		d.addToPack(s, SnowFallDensity(t))
	}

	d.adjustDegreeDayFactor(d.den)
	melt := d.ddf * df * (t - d.tb) // [m·°C-1·d-1]
	if melt > 0. {
		if melt >= d.swe-d.lwc {
			melt = d.swe - d.lwc
			d.internalFreeze(-melt)
			d.lwc = d.swe
		} else {
			d.internalFreeze(-melt)
			d.lwc += melt
		}
	} else {
		melt = 0.
	}

	if r > 0. {
		d.addToPack(r, pw)
		d.lwc += r
	}
	drainage = d.drainFromPack()

	// rfrz := ddrf * d.ddf * df * (d.tb - t) // [m·°C-1·d-1] // Seibert, J., 2005. HBV light version 2 User's Manual. Stockholm University Department of Physical Geography and Quaternary Geology, November, 2005. 32pp.
	// if rfrz < 0. {
	// 	rfrz = 0.
	// } else {
	// 	if rfrz > d.lwc {
	// 		rfrz = d.lwc
	// 	}
	// 	d.internalFreeze(rfrz)
	// 	d.lwc -= rfrz
	// }

	if d.swe <= 0. {
		d.swe = 0.
		d.ddf = ddfi // re-initialize ddf for adjustDegreeDayFactor()
	}
	return
}

// Properties returns the snowpack state
func (d *DDF) Properties() (porosity, depth, swe float64) {
	porosity, depth = d.properties()
	swe = d.swe
	return
}
