package swat

import "math"

const (
	twothird  = .66666666667
	fivethird = 1.66666666667
	secperday = 86400.
	zch       = 2. // SWAT default
	zfld      = 4. // SWAT default
)

// Channel is a channel units in SWAT
// ref: Neitsch, S.L., J.G. Arnold, J.R., Kiniry, J.R. Williams, 2011. Soil and Water Assessment Tool: Theoretical Documentation Version 2009 (September 2011). 647pp.
type Channel struct {
	len, sqlp, dbf, wbf, wbtm, wfld, zch float64 // geometry
	n, zch2, zfld2                       float64 // parameter
	vstr, d                              float64 // state variable (vstr: is the change in volume of storage during the time step m³)
}

// Route volumes (pg.432)
// variable storage rounting method (as built in HYMO)
// ref: Williams, J.R. and R.W. Hann, 1978. Optimal operation of large agricultural watersheds with water quality contraints. Texas Water Resources Institute, Texas A&M Univ. Tech. Rept. No. 96.
func (c *Channel) Route(vin float64) (vout float64) {
	a := c.vstr / 1000. / c.len // channel cross-sectional flow area [m²]
	p := func() float64 {       // wetted perimeter of the channel at depth [m] (pg.430-431)
		if c.d <= c.dbf {
			return c.wbtm + 2.*c.d*c.zch2
		}
		dfld := c.d - c.dbf
		return c.wbtm + 2.*c.dbf*c.zch2 + 4.*c.wbf + 2.*dfld*c.zfld2
	}()
	qout := math.Pow(a, fivethird) * c.sqlp / c.n / math.Pow(p, twothird) // qout,1 [m³/s]
	sc := 2. * secperday / (2.*c.vstr/qout + secperday)                   // sc: storage coefficient (pg.434); tt=Vstored,1/qout,1 travel time [s] pg.434

	vout = sc * (vin + c.vstr) // Vout,2 [m³]
	c.vstr -= vout             // Vstored,2
	a = c.vstr / 1000. / c.len // update FlowArea [m²]

	// update flowdepth [m]
	x := c.wbtm / 2. / c.zch
	c.d = math.Sqrt(a/c.zch+x*x) - x // pg.432
	if c.d > c.dbf {
		abf := (c.wbtm + c.zch*c.dbf) * c.dbf // FlowArea at bankful [m²] (pg.430)
		x := c.wfld / 2. / zfld
		c.d = c.dbf + math.Sqrt((a-abf)/zfld+x*x) - x
	}
	return
}
