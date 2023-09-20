package hechms

import (
	"fmt"

	"github.com/maseology/mmio"
)

func (m *Domain) Print(par Params) {

	bsn, _, totarea := m.initialize(&par)

	fmt.Printf("\nTotal Area: %.1f km2\n", totarea)
	lbsn := make([]string, len(bsn)+1)
	lbsn[0] = "name,ia,cn,pimp,tp,cp,k,rp,lag"
	for i, b := range bsn {
		ws := m.SBP[i]
		lbsn[i+1] = fmt.Sprintf("%s,%f,%f,%f,%f,%f,%f,%f,%f", ws.Name, b.ia, b.cn, b.fimp*100, b.tp, par.Cp, par.Kbf, b.rp, par.Krch*ws.FlowPathLen)
		// fmt.Println(lbsn[i+1])
	}
	mmio.WriteLines("BasinPrint.csv", lbsn)

	// lrch := make([]string, len(rch)+1)
	// lrch[0] = "name,ia,cn,pimp,tp,cp,k,rp"
	// // for i,  := range rch {
	// // 	ws := m.SBP[i]
	// // 	lbsn[i+1] = fmt.Sprintf("%s,%f,%f,%f,%f,%f,%f,%f", ws.Name, b.ia, b.cn, b.fimp*100, b.tp, par.Cp, par.Kbf, par.RatioToPeak)
	// // 	// fmt.Println(lbsn[i+1])
	// // }
	// mmio.WriteLines("ReachPrint.csv", lrch)
}
