package swat

import (
	"log"
	"math"
)

// Renew subbasin for resampling parameters
func (b *SubBasin) Renew(CNf, ESCO, CHN, OVN, SURLAG, GWDELAY, ALPHABF, GWQMN float64) {
	b.surlag = SURLAG // surface water lag coefficient
	b.dgw = GWDELAY   // the delay of soil zone percolation to aquifer [days]
	b.aqt = GWQMN     // the threshold water level in the shallow aquifer for groundwater contribution to the main channel to occur (mmH2O) (pg.175)
	b.agw = ALPHABF   // baseflow recession coefficient
	b.tribn = CHN     // effective Manning's n for subbasin tributaries

	// build HRUs and tconc, add channel element
	wslp, wovn := 0., 0.
	for _, m := range b.hru {
		cnadj := m.cn.cn
		if CNf > 0.5 {
			cnadj += 2. * (CNf - 0.5) * (99 - cnadj)
		} else {
			cnadj = 2.*CNf*(cnadj-1.) + 1.
		}

		m.cn.New(cnadj, m.sz[0].fc*float64(nsl), m.sz[0].sat*float64(nsl), m.slp) // fc, sat [mm]; SLP as fraction [m/m]
		m.ovn = OVN
		m.esco = ESCO
		wslp += m.f * m.slp // subsasin weighted average slope
		wovn += m.f * m.ovn // subsasin weighted average overland roughness
	}

	b.tconc = tconc(wslp, b.slplen, wovn, b.tribl, b.tribs, CHN, b.Ca)
	if math.IsNaN(b.tconc) {
		log.Fatalf("SubBasin.New error: tconc is NaN\n")
	}
}
