package rainrun

import (
	"math"

	"github.com/maseology/mmaths"
)

const timestep = 86400

type TOPMODEL struct {
	drel, nbins, kstb []float64
	d, fnc, m, n      float64
}

func (tm *TOPMODEL) New(drelative mmaths.Histogram, ksatTanBeta []float64, mTOPMODEL, porosity float64) {
	basinArea := 0.
	adrel, abins := make([]float64, len(drelative.Bins)), make([]float64, len(drelative.Bins))
	for i, n := range drelative.Bins {
		adrel[i] = drelative.Levels[i]
		abins[i] = float64(n)
		basinArea += float64(n)
	}
	tm.drel = adrel
	tm.nbins = abins
	tm.kstb = ksatTanBeta
	tm.fnc = basinArea
	tm.m = mTOPMODEL
	tm.n = porosity
}

func (tm *TOPMODEL) Update(p, ep float64) float64 {
	tm.d -= p
	q := 0.
	for i, dr := range tm.drel {
		di := tm.m*dr + tm.d
		if di < 0. {
			q += tm.kstb[i] * math.Exp(-di/tm.m) * tm.nbins[i]
		}
	}
	q *= timestep / tm.fnc
	tm.d += q
	return q // [m/ts]
}

func (tm *TOPMODEL) Storage() float64 {
	sdi := 0.
	for i, dr := range tm.drel {
		di := tm.m*dr + tm.d
		sdi += di * tm.nbins[i]
	}
	return sdi / tm.fnc
}

func (tm *TOPMODEL) DepthToWT(drelLocal float64) float64 {
	di := tm.m*drelLocal + tm.d
	return di / tm.n
}
