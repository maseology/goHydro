package hechms

import (
	"math"

	"github.com/maseology/goHydro/convolution"
)

func (m *Domain) initialize(par *Params) ([]basin, []reach, float64) {

	if len(m.Order) < 1 {
		m.Order = []int{0}
		m.MetXr = map[int]int{0: 0}
	}
	bsn := make([]basin, len(m.Order))
	rch := make([]reach, len(m.Order))
	totarea := 0.
	for _, i := range m.Order {
		w := m.SBP[i]
		// if _, ok := m.MetXr[w.MetID]; !ok {
		// 	panic("hechms.Domain.Run MetID error")
		// }
		if _, ok := m.MetXr[w.Swsid]; !ok {
			panic("hechms.Domain.Run Swsid (for MetID) error")
		}
		tp := .75 * par.Ct * math.Pow(w.FlowPathLen*w.CentFlowPathLen, .3)  // eq 6-6 [hr]
		trnfrm := convolution.Snyder2(w.Area, tp, par.Cp, float64(m.TSmin)) // []float64{1.} //
		ds := func() int {
			if len(m.Order) > 1 {
				if d, ok := m.Xr[w.Dsws]; ok {
					return d
				}
			}
			return -1
		}()
		// newcn := func() float64 {
		// 	sx := func(x float64) float64 {
		// 		return 1 / (1 + math.Exp(-4*math.Log10(x))) // to prevent CN>100
		// 	}
		// 	cn1 := sx(w.CN / 100. * par.Fcn)
		// 	if cn1 > 1 || cn1 < 0 {
		// 		panic("newcn")
		// 	}
		// 	if cn1 <= .01 { // CN=1 lower limit
		// 		cn1 = .01
		// 	}
		// 	return cn1 * 100.
		// }()
		newcn := math.Min(w.CN*par.Fcn, 99.)
		if par.Fcn <= 0 {
			newcn = w.CN
		}
		ia := par.Fia * w.Ia
		if par.Fia <= 0 {
			ia = w.Ia
		}
		bsn[i] = basin{
			ia:     ia, //par.Fia * w.Ia, // + par.Fia,
			trnfrm: trnfrm,
			qlag:   make([]float64, len(trnfrm)),
			cn:     newcn,
			scn:    25400./newcn - 254., // 25400./ w.CN / par.Fcn - 254., // mm
			area:   w.Area,
			fimp:   w.Fimp,
			peak:   -1,
			tfnext: -1,
			k:      math.Pow(par.Kbf, float64(m.TSmin)/60/24), // adjust baseflow exponential from per-day to per-timestep
			rp:     par.RatioToPeak,
			tp:     tp,
			dsid:   ds,
			mid:    m.MetXr[w.Swsid],
			// mid:    m.Mxr[w.MetID],
		}

		// simple lag
		rtrnfrm, r0 := convolution.DiracDelta(par.Krch*w.FlowPathLen, float64(m.TSmin))
		lag := make([]float64, len(rtrnfrm))
		rch[i] = &simplelag{
			trnfrm: rtrnfrm,
			lag:    lag,
			lag0:   r0,
		}
		// rch[i] = NewMuskingum(0., par.Krch, par.Q0, float64(m.TSmin)/60)

		totarea += w.Area
	}
	mm2cms := totarea * 1000. / 60. / float64(m.TSmin) // convert mm to cms
	for _, i := range m.Order {
		bsn[i].qbf = par.Q0 / mm2cms // convert cms to mm
	}

	return bsn, rch, totarea
}
