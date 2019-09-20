package hru

// // bit-wise status flag
// const (
// 	snowOnGround = 1 << iota
// 	waterOnSurface
// 	availPoreWater
// 	deficitPoreWater
// )

// WtrShd is an alias for a set of HRUs making up a watershed
type WtrShd = map[int]*HRU

// CopyWtrShd reutrns a deep copy of a WtrShd
func CopyWtrShd(origWtrShd WtrShd) (newWtrShd WtrShd) {
	newWtrShd = make(map[int]*HRU, len(origWtrShd))
	for k, v := range origWtrShd {
		newHRU := *v
		newWtrShd[k] = &newHRU
	}
	return
}

// HRU the Hydrologic Response Unit
type HRU struct {
	sma, srf         Res
	fimp, fprv, perc float64
	// stat             byte
}

// PercFimpCap returns percolation rates, fraction impervious, and storage capacity on the HRU
func (h *HRU) PercFimpCap() (perc, fimp, smacap, srfcap float64) {
	return h.perc, h.fimp, h.sma.cap, h.srf.cap
}

// Initialize HRU
func (h *HRU) Initialize(rzsto, srfsto, fimp, ksat, ts float64) {
	if rzsto < 0. || srfsto < 0. || fimp < 0. || fimp > 1. || ksat < 0. {
		panic("HRU Initialize parameter error")
	}
	h.sma.sto = 0.     // inital soil moisture storage
	h.srf.sto = 0.     // inital surface/depression storage
	h.sma.cap = rzsto  // soil moisture storage (i.e., rootzone/drainable storage)
	h.srf.cap = srfsto // surface/depression storage
	h.fimp = fimp      // fraction impervious
	h.fprv = 1. - fimp // fraction pervious
	// h.perc = ts * h.fprv * ksat // gravity-driven percolation rate m/ts
	h.perc = ts * ksat // gravity-driven percolation rate m/ts (unit gradient)
}

// Reset state
func (h *HRU) Reset() {
	h.sma.sto = 0. // inital storage
	h.srf.sto = 0. // inital storage
}

// Update hru given a set of forcings
func (h *HRU) Update(p, ep float64) (aet, ro, rch float64) {
	sri := h.fimp * p // impervious runoff
	ro = h.sma.Overflow(h.srf.Overflow(p-sri)) + sri
	rch = h.sma.Overflow(-h.perc) + h.perc
	avail := h.srf.Overflow(-ep) // remaining available ep
	avail = h.sma.Overflow(avail*h.fprv) + avail*h.fimp
	aet = ep + avail
	// h.updateStatus()
	return
}

// UpdateWT hru given a set of forcings and the presence of a watertable
func (h *HRU) UpdateWT(p, ep, zwt float64) (aet, ro, rch float64) {
	if zwt < 0. { // upward gradient
		rch = -h.sma.Deficit()           // groundwater discharge (negative recharge)
		h.sma.sto = h.sma.cap            // fill drainable porosity
		sri := h.fimp * p                // impervious runoff
		ro = h.srf.Overflow(p-sri) + sri // fulfill surface storage
		rch += h.srf.Overflow(-ep)       // remaining available ep assumed to be taken from high watertable
		aet = ep
	} else {
		aet, ro, rch = h.Update(p, ep)
	}
	return
}

// UpdateP adds precip to the hru and returns runoff
func (h *HRU) UpdateP(p float64) float64 {
	sri := h.fimp * p // impervious runoff
	return h.sma.Overflow(h.srf.Overflow(p-sri)) + sri
}

// UpdateEp removes evaporation from hru
func (h *HRU) UpdateEp(ep float64) float64 {
	avail := h.srf.Overflow(-ep) // remaining available ep
	avail = h.sma.Overflow(avail*h.fprv) + avail*h.fimp
	return ep + avail
}

// UpdatePerc updates hru given no forcings (percolation only)
func (h *HRU) UpdatePerc() float64 {
	return h.sma.Overflow(-h.perc) + h.perc // amount recharged
}

// UpdatePercWT updates hru over a high water table
// returns net groundwater exchange (negative: recharge, positive: discharge after filling soil zone)
// zwt: depth to watertable [m]
func (h *HRU) UpdatePercWT(zwt float64) float64 {
	if zwt < 0. { // upward gradient
		return h.sma.Overflow(h.perc * -zwt) // groundwater seepage, assumes unit gradient, unit thickness
	}
	return -(h.sma.Overflow(-h.perc) + h.perc) // returns the amount recharged
}

// UpdateIN updates hru given a set of input forcings only
func (h *HRU) UpdateIN(p float64) (ro, rch float64) {
	sri := h.fimp * p // impervious runoff
	ro = h.sma.Overflow(h.srf.Overflow(p-sri)) + sri
	rch = h.sma.Overflow(-h.perc) + h.perc
	// h.updateStatus()
	return
}

// UpdateStorage simply adds infiltration to storage (soil first, excess to surface depressions)
func (h *HRU) UpdateStorage(f float64) float64 {
	return h.srf.Overflow(h.sma.Overflow(f))
}

// AddToStorage simply adds infiltration to storage (soil first, excess to surface depressions), but keeps water onsite
func (h *HRU) AddToStorage(f float64) {
	// h.srf.sto += h.sma.Overflow(f)
	h.sma.sto += h.srf.Overflow(f)
}

// Storage returns total current storages
func (h *HRU) Storage() float64 {
	return h.sma.Storage() + h.srf.Storage()
}

// Deficit returns current storage deficit
func (h *HRU) Deficit() float64 {
	return h.sma.Deficit() + h.srf.Deficit()
}

// Infiltrability returns the amount of potential infiltration
func (h *HRU) Infiltrability() float64 {
	return (h.sma.Deficit() + h.srf.Deficit()) * h.fprv
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
