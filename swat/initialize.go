package swat

import (
	"fmt"
	"log"
	"math"

	"github.com/maseology/mmaths"
)

// Initialize SubBasin state
func (bsn *SubBasin) Initialize(q0mm float64) {
	bsn.qbf = q0mm
}

// Steady initializes model stores to a steady precipitation (input as mm/year)
func Steady(ws WaterShed, wf map[int]float64, ord []int, mmpy float64, msg bool) {
	const (
		tol = 1e-4 // convergence tolerance [mm]
		ef  = 1.   // ratio of ep to precip
	)
	p := mmpy / 365.24
	ep := ef * p

	stobsn := func(bsn *SubBasin) (aq, psto, qstr, vstr, sz float64) {
		aq = bsn.aq
		psto = bsn.psto
		qstr = bsn.qstr
		vstr = bsn.chn.vstr / bsn.Ca / 1000.
		sz = 0.
		for _, m := range bsn.hru {
			sz += m.storage() * m.f
		}
		return
	}

	stows := func() (aq, psto, qstr, vstr, sz float64) {
		aq, psto, qstr, vstr, sz = 0., 0., 0., 0., 0.
		for _, sbid := range ord {
			aqt, pstot, qstrt, vstrt, szt := stobsn(ws[sbid])
			aq += aqt * wf[sbid]
			psto += pstot * wf[sbid]
			qstr += qstrt * wf[sbid]
			vstr += vstrt * wf[sbid]
			sz += szt * wf[sbid]
		}
		return
	}

	var pp, r, i, a, g, b float64
	aq0, psto0, qstr0, vstr0, sz0 := stows()
	c := 0
	for {
		c++
		vin := make(map[int]float64, len(ws))
		pp, r, i, a, g, b = 0., 0., 0., 0., 0., 0.
		for _, sbid := range ord {
			bsn := ws[sbid]
			var rt, it, at, gt, bt, vout float64
			if _, ok := vin[sbid]; ok {
				rt, it, at, gt, bt, vout = bsn.Update(vin[sbid], p, ep)
			} else {
				rt, it, at, gt, bt, vout = bsn.Update(0., p, ep)
			}
			if bsn.Outflow >= 0. {
				vin[bsn.Outflow] += vout
			}
			pp += p * wf[sbid]
			r += rt * wf[sbid]
			i += it * wf[sbid]
			a += at * wf[sbid]
			g += gt * wf[sbid]
			b += bt * wf[sbid]
		}
		aq1, psto1, qstr1, vstr1, sz1 := stows()
		diff := mmaths.RelativeDifference(aq1, aq0)
		diff += mmaths.RelativeDifference(psto1, psto0)
		diff += mmaths.RelativeDifference(qstr1, qstr0)
		diff += mmaths.RelativeDifference(vstr1, vstr0)
		diff += mmaths.RelativeDifference(sz1, sz0)
		if math.IsNaN(diff) {
			log.Fatalf("SubBasin.Steady error, diff is NaN")
		}
		if math.Abs(diff) < tol {
			break
		}
		aq0, psto0, qstr0, vstr0, sz0 = aq1, psto1, qstr1, vstr1, sz1
	}

	if msg {
		fmt.Printf("watershed converged in %d iterations\n", c)
		f := 365.24
		pp *= f
		r *= f // [mm/yr]
		i *= f // [mm/yr]
		a *= f // [mm/yr]
		g *= f // [mm/yr]
		b *= f // [mm/yr]
		fmt.Printf(" precip: %.3f runoff: %.3f infiltration: %.3f aet: %.3f recharge: %.3f baseflow: %.3f\n", pp, r, i, a, g, b)
		print("")
	}
}
