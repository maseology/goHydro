package convolution

import "math"

type RatingCurve struct{ Q, A, W, H []float64 }

const (
	hmax  = 10.  // [m]
	hstep = .001 // [m]
)

func round3(v float64) float64 {
	return math.Round(v*1000.) / 1000.
}

// NewTrapezoid creates a rating curve for a Trapezoidal channel
//
// Cd weir coefficient; b base width; t side slope (t horizontal to 1 vertical)
//
// from Gupta (1989) Hydrology and Hydraulic Systems pg.267
func NewTrapezoid(Cd, b, t float64) *RatingCurve {
	i, n := 0, int(hmax/hstep)
	twothirds := 2. / 3.
	eightfifteehths := 8. / 15.
	sqrt2g := math.Sqrt(2 * 9.8067)

	aq, aa, aw, ah := make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n)
	for h := 0.; h < hmax; h += hstep {
		h = round3(h)
		ah[i] = h             // depth
		aa[i] = (b + t*h) * h // area
		aw[i] = b + 2*t*h     // top width
		// P := b + 2*h*math.Sqrt(1+t*t) // wetted perimeter
		// R := aa[i] / P                // hydraulic radius
		// D := aa[i] / aw[i]            // hydraulic depth
		aq[i] = twothirds*Cd*sqrt2g*(b-.2*h)*math.Pow(h, 1.5) + eightfifteehths*Cd*sqrt2g*t*math.Pow(h, 2.5) // rectangle + triangle
		i++
	}
	return &RatingCurve{aq, aa, aw, ah}
}

// NewRectangular creates a rating curve for a Rectangular channel
//
// Cd weir coefficient; b base width
//
// from Gupta (1989) Hydrology and Hydraulic Systems pg.267
func NewRectangular(Cd, b float64) *RatingCurve {
	i, n := 0, int(hmax/hstep)
	twothirds := 2. / 3.
	sqrt2g := math.Sqrt(2 * 9.8067)

	aq, aa, aw, ah := make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n)
	for h := 0.; h < hmax; h += hstep {
		ah[i] = h       // depth
		aa[i] = b * h   // area
		aw[i] = b + 2*h // top width
		// P := b + 2*h       // wetted perimeter
		// R := aa[i] / P     // hydraulic radius
		// D := aa[i] / aw[i] // hydraulic depth
		aq[i] = twothirds * Cd * sqrt2g * (b - .2*h) * math.Pow(h, 1.5)
		i++
	}

	println(ah[n-1])

	return &RatingCurve{aq, aa, aw, ah}
}

// NewTriangle creates a rating curve for a Triangular weir
//
// Cd weir coefficient; t side slope (t horizontal to 1 vertical)
//
// from Gupta (1989) Hydrology and Hydraulic Systems pg.267
func NewTriangle(Cd, t float64) *RatingCurve {
	i, n := 0, int(hmax/hstep)
	eightfifteehths := 8. / 15.
	sqrt2g := math.Sqrt(2 * 9.8067)

	aq, aa, aw, ah := make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n)
	for h := 0.; h < hmax; h += hstep {
		ah[i] = h         // depth
		aa[i] = t * h * h // area
		aw[i] = 2 * t * h // top width
		// P := 2 * h * math.Sqrt(1+t*t)                                // wetted perimeter
		// R := aa[i] / P                                               // hydraulic radius
		// D := aa[i] / aw[i]                                           // hydraulic depth
		aq[i] = eightfifteehths * Cd * sqrt2g * t * math.Pow(h, 2.5)
		i++
	}

	println(ah[n-1])

	return &RatingCurve{aq, aa, aw, ah}
}

// hbank, wfloodplain in [m]
func NewCompoundTrapezoid(Cd, b, t, hbank, wfloodplain float64) *RatingCurve {
	i, n := 0, int(hmax/hstep)
	sqrt2g := math.Sqrt(2 * 9.8067)
	f1 := 2. / 3. * Cd * sqrt2g
	f2 := 8. / 15. * Cd * sqrt2g * t

	aq, aa, aw, ah := make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n)
	for h := 0.; h < hmax; h += hstep {
		h = round3(h)
		ah[i] = h // depth
		if h <= hbank {
			aa[i] = (b + t*h) * h // area
			aw[i] = b + 2*t*h     // top width
			// P := b + 2*h*math.Sqrt(1+t*t) // wetted perimeter
			// R := aa[i] / P                // hydraulic radius
			// D := aa[i] / aw[i]            // hydraulic depth
			aq[i] = f1*(b-.2*h)*math.Pow(h, 1.5) + f2*math.Pow(h, 2.5) // rectangle + triangle
		} else {
			// main channel
			aa[i] = (b + t*hbank) * hbank                                          // area
			aq[i] = f1*(b-.2*hbank)*math.Pow(hbank, 1.5) + f2*math.Pow(hbank, 2.5) // rectangle + triangle

			// compound
			hcomp := h - hbank
			aa[i] += (wfloodplain + t*hcomp) * hcomp                                          // area
			aq[i] += f1*(wfloodplain-.2*hcomp)*math.Pow(hcomp, 1.5) + f2*math.Pow(hcomp, 2.5) // rectangle + triangle
			aw[i] = wfloodplain + 2*t*hcomp                                                   // top width
		}
		i++
	}
	return &RatingCurve{aq, aa, aw, ah}
}
