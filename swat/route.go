package swat

import (
	"log"
	"math"
)

const (
	twothird = 2. / 3.
	zch      = 2. // inverse of the channel side slope (SWAT default pg.429)
	zfld     = 4. // SWAT default
)

// Route volumes (pg.432)
// variable storage rounting method (as built in HYMO)
// ref: Williams, J.R. and R.W. Hann, 1978. Optimal operation of large agricultural watersheds with water quality contraints. Texas Water Resources Institute, Texas A&M Univ. Tech. Rept. No. 96.
// ref: Williams J.R., 1969. Flood routing with variable travel time or variable storage coefficients. Transactions of the ASAE 12(1): 100--103.
func (c *Channel) Route(vinavg float64) (vout2 float64) {
	if c.len == 0. { // no channel, direct translation
		return vinavg
	}
	vsv := c.vstr
	c.vstr += vinavg
	vout2 = c.sc * c.vstr
	c.vstr -= vout2 // computing Vsto,2 from Vsto,1 - Vout,2
	if c.vstr < 0. {
		log.Fatalf("Channel.Route error: c.vstr < 0: %f\n", c.vstr)
	}
	if vout2 == 0. {
		return
	}

	// // update storage coefficient (pg.434)
	// tt := c.vstr * secperday / vout2 // travel time [s]
	// if tt < secperday/2. {
	// 	tt = secperday / 2.
	// }
	// c.sc = 2. * secperday / (2.*tt + secperday)

	wbal := c.vstr - vsv + vout2 - vinavg
	if math.Abs(wbal) > 1e-6 {
		log.Fatalf("Channel.Route error: |wbal| = %f\n", wbal)
	}
	if vinavg > 0 && vout2 == 0 {
		log.Fatalf("Channel.Route error: vinavg > 0 && vout2 == 0\n")
	}
	return

	// ach := c.vstr / c.len // channel cross-sectional flow area [m²]
	// p := func() float64 { // wetted perimeter of the channel at depth [m] (pg.430-431)
	// 	if c.d <= c.dbf {
	// 		return c.wbtm + 2.*c.d*c.zch2
	// 	}
	// 	dfld := c.d - c.dbf
	// 	return c.wbtm + 2.*c.dbf*c.zch2 + 4.*c.wbf + 2.*dfld*c.zfld2
	// }()
	// qout := ach * math.Pow(ach/p, twothird) * c.sqslp / c.n // qout,1 [m³/s]
	// sc := 2. * secperday / (2.*c.vstr/qout + secperday)     // sc: storage coefficient (pg.434); tt=Vstored,1/qout,1 travel time [s] pg.434

	// c.vstr += vinavg     // add upstream sources
	// vout2 = sc * c.vstr   // Vout,2 [m³] pg.434
	// c.vstr -= vout2       // Vstored,2
	// ach = c.vstr / c.len // update FlowArea [m²]

	// // update flowdepth [m]
	// x := c.wbtm / 2. / c.zch
	// c.d = math.Sqrt(ach/c.zch+x*x) - x // pg.432
	// if c.d > c.dbf {
	// 	abf := (c.wbtm + c.zch*c.dbf) * c.dbf // FlowArea at bankful [m²] (pg.430)
	// 	x := c.wfld / 2. / zfld
	// 	c.d = c.dbf + math.Sqrt((ach-abf)/zfld+x*x) - x
	// }
	// return
}
