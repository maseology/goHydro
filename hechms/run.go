package hechms

import (
	"fmt"
	"time"

	"github.com/maseology/goHydro/convolution"
	"github.com/maseology/goHydro/forcing"
)

const tsmin = 15

func (m *Domain) Run(frc *forcing.Forcing, dtb, dte time.Time, par Params) ([]time.Time, []float64, []float64) {

	ste := make([]state, len(m.Order))
	totarea := 0.
	for _, i := range m.Order {
		w := m.SB[i]
		if _, ok := m.Mxr[w.MetID]; !ok {
			panic("hechms.Domain.Run MetID error")
		}
		trnfrm := convolution.Snyder2(w.Area, par.Tb, par.Cp, tsmin)
		ste[i] = state{
			ia:     0., // w.Percov,
			trnfrm: trnfrm,
			qlag:   make([]float64, len(trnfrm)),
			cn:     w.CN,
			area:   w.Area,
			fimp:   w.Perimp,
			mid:    m.Mxr[w.MetID],
		}
		totarea += w.Area
	}
	totarea /= 1000. // to convert outputs to mm

	jtb, jte := func() (int, int) {
		j0, j1 := -1, -1
		for i, t := range frc.T {
			if j0 < 0 && t.After(dtb) {
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
	ns := int(substeps) * (jte - jtb + 1)
	sim, pre := make([]float64, ns), make([]float64, ns)
	dts := make([]time.Time, ns)
	pcum, qcum := 0., 0.
	for j := jtb; j <= jte; j++ {
		for k := 0; k < substeps; k++ {
			jj := (j-jtb)*substeps + k
			dts[jj] = frc.T[j].Add(time.Second * time.Duration(timestep*k))
			qall, psum := 0., 0.
			for _, i := range m.Order {
				p, q := frc.Ya[ste[i].mid][j]/float64(substeps), 0.
				if p > 0. {
					q = ste[i].scscn(p*(1-ste[i].fimp)) + p*ste[i].fimp // Loss
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
	fmt.Println(pcum, qcum, qcum/pcum)
	return dts, sim, pre
}
