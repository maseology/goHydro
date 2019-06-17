package swat

import "math"

const (
	twothird  = 2. / 3.
	fivethird = 5. / 3.
	secperday = 86400.
	zch       = 2. // SWAT default
	zfld      = 4. // SWAT default
)

// Route volumes (pg.432)
// variable storage rounting method (as built in HYMO)
// ref: Williams, J.R. and R.W. Hann, 1978. Optimal operation of large agricultural watersheds with water quality contraints. Texas Water Resources Institute, Texas A&M Univ. Tech. Rept. No. 96.
// ref: Williams J.R., 1969. Flood routing with variable travel time or variable storage coefficients. Transactions of the ASAE 12(1): 100--103.
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

	c.vstr += vin
	vout = sc * c.vstr         // Vout,2 [m³] pg.434
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
