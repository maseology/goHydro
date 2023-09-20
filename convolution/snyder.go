package convolution

import (
	"math"

	"github.com/maseology/mmaths"
)

// Snyder returns the oridinates for a 7-point Snyder (1938) unit hydrograph
// see pg.118 of Bedient, P.B., W.C. Huber, 2002. Hydrology and Floodplain Analysis, 3rd ed. Prentice Hall. 763pp.
// ref: Snyder, F.F., 1938. Synthetic Unit Hydrograph. Trans. AGU, vol.19. pp. 447-454.
func Snyder(Akm2, Ct, Cp, L, Lc, tsminutes float64) []float64 {
	kmTomi := func(km float64) float64 {
		return km / 1.60934
	}
	tp := Ct * math.Pow(kmTomi(L)*kmTomi(Lc), .3) // hr
	return Snyder2(Akm2, tp, Cp, tsminutes)
}

// Snyder2 same as above, only lag time and peak coefficient are specified (like HEC-HMS)
// Cp=[.4,.8]
func Snyder2(Akm2, tp, Cp, tsminutes float64) []float64 {
	// following example pg. 119 Cebient Huber
	ami2 := Akm2 / 1.60934 / 1.60934      // convert to mi²
	qp := 640 * Cp * ami2 / tp            // cfs
	tb := 4 * tp                          // hr; for small watersheds
	w75 := 440 * math.Pow(qp/ami2, -1.08) // widths are distributed 1/3 before Qp and 2/3 after
	w50 := 720 * math.Pow(qp/ami2, -1.08)

	pts := []mmaths.Point{
		{X: 0, Y: 0},
		{X: tp - w50/3, Y: qp / 2},
		{X: tp - w75/3, Y: 3 * qp / 4},
		{X: tp, Y: qp},
		{X: tp + 2*w75/3, Y: 3 * qp / 4},
		{X: tp + 2*w50/3, Y: qp / 2},
		{X: tb, Y: 0},
	}
	for _, c := range pts {
		if c.X < 0 {
			panic("Snyder2: negative ordinate")
		}
	}
	tshr := tsminutes / 60.
	nstep := int(math.Floor(tb / tshr))
	return buildOrdinatesFromPoints(pts, tshr, nstep)
}
