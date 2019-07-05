package swat

import "math"

// surfRunoffLag returns the the amount of surface runoff released to the main channel (pg.116)
func (bsn *SubBasin) surfRunoffLag(qgen float64) float64 {
	qsurf := (qgen + bsn.qstr) * (1. - math.Exp(-bsn.surlag/bsn.tconc))
	bsn.qstr += qgen - qsurf // update state
	return qsurf
}
