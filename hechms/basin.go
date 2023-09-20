package hechms

type basin struct {
	trnfrm, qlag                                                    []float64
	Pe, F, Q, qbf, ia, cn, scn, area, fimp, peak, tfnext, k, rp, tp float64 // Pe, F, Q are cumulative effective precip, infiltration and runoff
	mid, dsid                                                       int
}
