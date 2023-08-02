package hechms

import (
	"fmt"
	"time"

	"github.com/maseology/goHydro/convolution"
	"github.com/maseology/goHydro/forcing"
)

const tsmin = 15

func (m *Domain) Run(frc *forcing.Forcing, dtb, dte time.Time, par Params) ([]time.Time, []float64, []float64) {

	if len(m.Order) < 1 {
		m.Order = []int{0}
		m.Mxr = map[int]int{0: 0}
	}
	ste := make([]state, len(m.Order))
	totarea := 0.
	for _, i := range m.Order {
		w := m.SB[i]
		// if _, ok := m.Mxr[w.MetID]; !ok {
		// 	panic("hechms.Domain.Run MetID error")
		// }
		if _, ok := m.Mxr[w.Swsid]; !ok {
			panic("hechms.Domain.Run Swsid (for MetID) error")
		}
		trnfrm := convolution.Snyder2(w.Area, par.Tb, par.Cp, tsmin) // []float64{1.} //
		ste[i] = state{
			ia:     w.Ia,
			trnfrm: trnfrm,
			qlag:   make([]float64, len(trnfrm)),
			scn:    25400./w.CN - 254., // mm
			area:   w.Area,
			fimp:   w.Fimp,
			// mid:    m.Mxr[w.MetID],
			mid: m.Mxr[w.Swsid],
		}
		totarea += w.Area
	}

	jtb, jte := func() (int, int) { // get interval by index
		j0, j1 := -1, -1
		for i, t := range frc.T {
			if j0 < 0 && t.After(dtb) || t.Equal(dtb) {
				j0 = i
			}
			if j0 >= 0 && t.After(dte) {
				j1 = i
				break
			}
		}
		if j1 < 0 {
			j1 = len(frc.T) - 1
		}
		return j0, j1
	}()

	timestep := tsmin * 60
	substeps := int(frc.IntervalSec) / timestep
	ns, fss := substeps*(jte-jtb+1), float64(substeps)
	sim, pre := make([]float64, ns), make([]float64, ns)
	dts := make([]time.Time, ns)
	pcum, qcum := 0., 0.
	for j := jtb; j <= jte; j++ {
		for k := 0; k < substeps; k++ {
			jj := (j-jtb)*substeps + k
			dts[jj] = frc.T[j].Add(time.Second * time.Duration(timestep*k))
			qall, psum := 0., 0.
			for _, i := range m.Order {
				p, q := frc.Ya[ste[i].mid][j]/fss, 0.
				if p > 0. {
					q = ste[i].scscn(p)*(1-ste[i].fimp) + p*ste[i].fimp // Loss
					for j, u := range ste[i].trnfrm {
						ste[i].qlag[j] += q * u // transform
					}
				}
				qall += ste[i].qlag[0] * ste[i].area
				ste[i].qlag = append(ste[i].qlag[1:], 0.)
				psum += p * ste[i].area
			}
			sim[jj] = qall / totarea
			pre[jj] = psum / totarea
			qcum += sim[jj]
			pcum += pre[jj]
		}
	}
	fmt.Printf(" p: %.1f q: %.1f mm  q/p = %.3f  qmax: %.1f cms\n", pcum, qcum, qcum/pcum, func() float64 {
		mx := 0.
		for _, m := range sim {
			if m > mx {
				mx = m
			}
		}
		return mx * totarea * 1000. / 60. / float64(tsmin) // convert to cms
	}())
	// for j := range sim {
	// 	sim[j] *= totarea * 1000. / 60. / float64(tsmin) // convert to cms
	// }
	return dts, sim, pre
}
