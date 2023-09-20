package hechms

import (
	"github.com/maseology/goHydro/forcing"
	"github.com/maseology/goHydro/hyetograph"
)

func (m *Domain) Run(frc *forcing.Forcing, jtb, jte, offset int, par Params) ([]float64, []float64) {

	bsn, rch, totarea := m.initialize(&par)
	return m.run(frc, bsn, rch, totarea, jtb, jte, offset)

}

func (m *Domain) run(frc *forcing.Forcing, bsn []basin, rch []reach, totarea float64, jtb, jte, offset int) ([]float64, []float64) {

	mm2cms := totarea * 1000. / 60. / float64(m.TSmin) // convert mm to cms
	timestep := m.TSmin * 60
	substeps := int(frc.IntervalSec) / timestep
	ns := substeps * (jte - jtb + 1)
	// fss := float64(substeps)
	yf := hyetograph.Unit(substeps) //hyetograph.SCSII(substeps, 6) //
	sim, pre := make([]float64, ns), make([]float64, ns)
	pcum, qcum := 0., 0.
	for j := jtb; j <= jte; j++ {
		for k := 0; k < substeps; k++ {
			jj := (j-jtb)*substeps + k
			qall, psum := 0., 0.
			for _, i := range m.Order {
				// p, q := frc.Ya[bsn[i].mid][j]/fss, 0.
				p, q := yf[k]*frc.Ya[bsn[i].mid][j+offset], 0.
				if p > 0. {
					q = bsn[i].scscn(p)*(1-bsn[i].fimp) + p*bsn[i].fimp // Loss
					for v, u := range bsn[i].trnfrm {
						bsn[i].qlag[v] += q * u // direct flow to transform
					}
				}
				df := bsn[i].qlag[0] // pop "direct flow"
				bsn[i].qlag = append(bsn[i].qlag[1:], 0.)
				if df < 0 {
					panic("negative direct flow")
				}

				tf := df + bsn[i].qbf // "total flow" [mm]
				if df == 0. {
					bsn[i].peak = -1.   // reset storm
					bsn[i].tfnext = -1. // disable special case
					bsn[i].qbf *= bsn[i].k
				} else if bsn[i].tfnext > 0 && df < bsn[i].tfnext { // special case: df>0, but in recession period
					tf = bsn[i].tfnext
					bsn[i].tfnext *= bsn[i].k
					bsn[i].qbf = tf - df
				} else { // event
					if tf > bsn[i].peak {
						bsn[i].peak = tf
					}
					if tf < bsn[i].peak*bsn[i].rp {
						bsn[i].qbf += bsn[i].peak*bsn[i].rp - tf
						tf = bsn[i].peak * bsn[i].rp
						bsn[i].tfnext = tf * bsn[i].k
					}
					bsn[i].qbf *= bsn[i].k
				}

				tf *= bsn[i].area // (volumetric) flow [mm.km2]
				qall += tf
				psum += p * bsn[i].area

				// reach routing
				tf += rch[i].Update(-1)
				di := bsn[i].dsid
				if di < 0 { // farfield
					sim[jj] += tf // [mm.km2] (assumes models with only 1 output)
				} else {
					rch[di].Update(tf)
				}
			}
			// sim[jj] = qall / totarea
			pre[jj] = psum / totarea // [mm]
			sim[jj] /= totarea       // [mm]
			qcum += sim[jj]
			pcum += pre[jj]
		}
	}
	// fmt.Printf(" p: %.1f q: %.1f mm  q/p = %.3f  qmax: %.1f cms\n", pcum, qcum, qcum/pcum, func() float64 {
	// 	mx := 0.
	// 	for _, m := range sim {
	// 		if m > mx {
	// 			mx = m
	// 		}
	// 	}
	// 	return mx * mm2cms
	// }())
	for j := range sim {
		sim[j] *= mm2cms
	}
	return sim, pre
}
