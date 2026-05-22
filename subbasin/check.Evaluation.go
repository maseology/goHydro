package subbasin

import "github.com/maseology/goHydro/grid"

func CheckAndPrint(ev *Evaluator, gd *grid.Definition, cids, igw []int, chkdirprfx string, crop bool) {

	var gdcrp *grid.Definition
	xr := make(map[int]int)
	if crop {
		gdcrp, xr = gd.CropToActives()
	} else {
		gdcrp = gd
		for _, c := range gd.Sactives {
			xr[c] = c
		}
	}
	_ = gdcrp

	// // output
	// sgw := gdcrp.NullInt32(-9999)
	// drel, bo, fcasc, finf, dsto, m := gdcrp.NullArray(-9999.), gdcrp.NullArray(-9999.), gdcrp.NullArray(-9999.), gdcrp.NullArray(-9999.), gdcrp.NullArray(-9999.), gdcrp.NullArray(-9999.)
	// for k, aids := range ev.Sais {
	// 	for i, ac := range aids {
	// 		var c int
	// 		if x, ok := xr[cids[ac]]; ok {
	// 			c = x
	// 		} else {
	// 			fmt.Println(ac, cids[ac])
	// 			panic("rdrr.Evaluator CheckAndPrint 1")
	// 		}
	// 		// c := xr[cids[ac]]
	// 		drel[c] = ev.Drel[k][i]
	// 		bo[c] = ev.Bo[k][i]
	// 		fcasc[c] = ev.Fcasc[k][i]
	// 		finf[c] = ev.Finf[k][i]
	// 		dsto[c] = ev.DepSto[k][i]
	// 		m[c] = ev.M[igw[ac]]
	// 		sgw[c] = int32(ev.Sgw[k])
	// 		// istrm[c] = func() int32 {
	// 		// 	if ev.IsStrm[k][i] {
	// 		// 		return 1
	// 		// 	}
	// 		// 	return 0
	// 		// }()
	// 	}
	// }

	// writeInts(gdcrp, chkdirprfx+"evaluator.sgw.bil", sgw) // groundwater index, now projected to sws
	// // writeInts(gdcrp, chkdirprfx+"evaluator.isstrm.bil", istrm)     // stream cells
	// writeFloats32(gdcrp, chkdirprfx+"evaluator.drel.bil", drel)    // groundwater deficit relative to the regional mean (deltaD)
	// writeFloats32(gdcrp, chkdirprfx+"evaluator.bo.bil", bo)        // groundwater flux to surface/channels
	// writeFloats32(gdcrp, chkdirprfx+"evaluator.fcasc.bil", fcasc)  // fraction of excess storage to runoff
	// writeFloats32(gdcrp, chkdirprfx+"evaluator.finf.bil", finf)    // fraction of excess storage to infiltrate assuming a falling head through a unit length per timestep
	// writeFloats32(gdcrp, chkdirprfx+"evaluator.depsto.bil", dsto)  // depression storage
	// writeFloats32(gdcrp, chkdirprfx+"evaluator.TOPMODEL-m.bil", m) // TOPMODEL parameter m
}
