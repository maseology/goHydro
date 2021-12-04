package pet

import (
	"math"
	"time"
)

// SineCurve returns (0.,1.) with summer solstice at 1.
func SineCurve(dt time.Time) float64 {
	const offset_days = 15.
	doy := float64(dt.YearDay()) - offset_days
	return .5 * (1. - math.Cos(2.*math.Pi*doy/365.24))
}

func SineCurvePET(AnnualPET float64, dt time.Time) float64 {
	const pet_min = 0.
	return SineCurve(dt)*(AnnualPET/365.24-pet_min) + pet_min
}
