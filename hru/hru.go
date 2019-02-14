package hru

// // bit-wise status flag
// const (
// 	snowOnGround = 1 << iota
// 	waterOnSurface
// 	availPoreWater
// 	deficitPoreWater
// )

// Basin is an alias for a set of HRUs
type Basin = map[int]HRU

// HRU the Hydrologic Response Unit
type HRU struct {
	sma, srf         res
	fimp, fprv, perc float64
	// stat             byte
}

// Initialize HRU
func (h *HRU) Initialize(rzsto, srfsto, fimp, ksat, ts float64) {
	if rzsto < 0. || srfsto < 0. || fimp < 0. || fimp > 1. || ksat < 0. {
		panic("HRU Initialize parameter error")
	}
	h.sma.sto = 0.              // inital storage
	h.srf.sto = 0.              // inital storage
	h.sma.cap = rzsto           // soil moisture storage (i.e., rootzone/drainable storage)
	h.srf.cap = srfsto          // surface/depression storage
	h.fimp = fimp               // fraction impervious
	h.fprv = 1. - fimp          // fraction pervious
	h.perc = ts * h.fprv * ksat // gravity-driven percolation rate m/ts
}

// Update hru given a set of forcings
func (h *HRU) Update(p, ep float64) (aet, ro, rch float64) {
	sri := h.fimp * p // impervious runoff
	ro = h.sma.overflow(h.srf.overflow(p-sri)) + sri
	rch = h.sma.overflow(-h.perc) + h.perc
	avail := h.srf.overflow(-ep) // remaining available ep
	avail = h.sma.overflow(avail*h.fprv) + avail*h.fimp
	aet = ep + avail
	// h.updateStatus()
	return
}

// Storage returns total current storages
func (h *HRU) Storage() float64 {
	return h.sma.Storage() + h.srf.Storage()
}

// func (h *HRU) updateStatus() {
// 	if h.sma.sto < h.sma.cap {
// 		h.stat |= deficitPoreWater
// 		if h.sma.sto > 0. {
// 			h.stat |= availPoreWater
// 		} else {
// 			h.stat &^= availPoreWater
// 		}
// 	} else {
// 		h.stat &^= deficitPoreWater
// 		if h.sma.sto > h.sma.cap {
// 			h.stat |= waterOnSurface
// 		} else {
// 			h.stat &^= waterOnSurface
// 		}
// 	}
// }
