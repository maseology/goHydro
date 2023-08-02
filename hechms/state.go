package hechms

type state struct {
	trnfrm, qlag                  []float64
	Pe, F, Q, ia, scn, area, fimp float64 // Pe, F, Q are cumulative effective precip, infiltration and runoff
	mid                           int
}
