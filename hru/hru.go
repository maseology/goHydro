package hru

import (
	"log"
	"math"
)

// // bit-wise status flag
// const (
// 	snowOnGround = 1 << iota
// 	waterOnSurface
// 	availPoreWater
// 	deficitPoreWater
// )

// WtrShd is an alias for a set of HRUs making up a watershed
type WtrShd = map[int]*HRU

// CopyWtrShd returns a deep copy of a WtrShd
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
	Sma, Sdet  Res     // retention reservoir Sh (where water has the potential to drain); detention reservoir Sk (where water is held locally)
	Fimp, Perc float64 // Fraction impervious; percolation rates
	// Fimp, Fprv, Perc float64
	// stat             byte // status
}

// Storage returns total current storages
func (h *HRU) Storage() float64 {
	return h.Sma.Sto + h.Sdet.Sto
}

// // PercFimpCap returns percolation rates, fraction impervious, and storage capacity on the HRU
// func (h *HRU) PercFimpCap() (perc, fimp, smacap, srfcap float64) {
// 	return h.Perc, h.Fimp, h.Sma.Cap, h.Sdet.Cap
// }

// Initialize HRU
func (h *HRU) Initialize(rzsto, srfsto, fimp, ksat, sma0, srf0 float64) {
	if rzsto < 0. || srfsto < 0. || fimp < 0. || fimp > 1. || ksat < 0. || sma0 < 0. || srf0 < 0. {
		log.Fatalf("HRU Initialize parameter error: rzsto=%.5f srfsto=%.5f fimp=%.5f ksat=%.5f\n", rzsto, srfsto, fimp, ksat)
	}
	// if ksat > rzsto {
	// 	fmt.Printf("HRU Initialize parameter warning: ksat > rzsto; percolation will never exceed drainable storage")
	// }
	h.Sma.Sto = sma0    // initial soil moisture storage
	h.Sdet.Sto = srf0   // initial surface/depression storage
	h.Sma.Cap = rzsto   // soil moisture storage (i.e., rootzone/drainable storage)
	h.Sdet.Cap = srfsto // surface/depression storage
	h.Fimp = fimp       // fraction impervious
	// h.Fprv = 1. - fimp  // fraction pervious
	h.Perc = ksat // gravity-driven percolation rate m/ts (unit gradient)
	// h.Perc =  h.Fprv * ksat // gravity-driven percolation rate m/ts

}

// Reset state
func (h *HRU) Reset() {
	h.Sma.Sto = 0.  // inital storage
	h.Sdet.Sto = 0. // inital storage
}

// Update hru given a set of forcings
func (h *HRU) Update(p, ep float64) (aet, ro, rch float64) {
	rp := h.Sdet.Overflow(p)          // flush detention storage
	sri := h.Fimp * rp                // impervious runoff
	ro = h.Sma.Overflow(rp-sri) + sri // flush retention, compute potential runoff
	if ep > 0 {
		avail := h.Sdet.Overflow(-ep) // remove ep from detention
		// f := (1. - h.Fimp) * h.Sma.Deficit()           // note: h.Sma has been flushed
		// avail = h.Sma.Overflow(avail*f) + avail*(1.-f) // remaining available ep (cannot be >0.)
		avail = h.Sma.Overflow(avail*(1.-h.Fimp)) + avail*h.Fimp // remaining available ep (cannot be >0.)
		aet = ep + avail                                         // actual et
	}
	rch = h.Sma.Overflow(-h.Perc) + h.Perc // compute total water percolated
	// h.updateStatus()
	return
}

// // UpdateWT hru given a set of forcings and the presence of a high watertable
// func (h *HRU) UpdateWT(p, ep float64, upwardGradient bool) (aet, ro, rch float64) {
// 	if upwardGradient {
// 		x := h.Sma.Sto - h.Sma.Cap // excess stored (drainable)
// 		gwd := 0.
// 		if x < 0. { // fill remaining deficit, assume discharge
// 			gwd = x // groundwater discharge (negative recharge)
// 			x = 0.
// 		}
// 		h.Sma.Sto = h.Sma.Cap // saturate retention reservoir (drainable porosity)

// 		aet, ro, rch = h.Update(p+x, ep)
// 		rch += gwd

// 		// x := h.Sma.Sto - h.Sma.Cap // excess stored (drainable)
// 		// if x < 0. { // fill remaining deficit, assume discharge
// 		// 	rch = x
// 		// 	x = 0.
// 		// }
// 		// h.Sma.Sto = h.Sma.Cap // saturate retention reservoir (drainable porosity)

// 		// ro = h.Sdet.Overflow(p) + x   // fulfill detention reservoir, add excess to runoff
// 		// avail := h.Sdet.Overflow(-ep) // remove ep from detention
// 		// // option 1 no evap from gw
// 		// aet = ep + avail // actual et
// 		// // // // option 2 unlimited evap from gw
// 		// // // rch += avail // ep assumed unlimited from a saturated surface (Note: avail cannot be >0.)
// 		// // // aet = ep     // completely satisfied over a high watertable
// 		// // // option 3 limited evap
// 		// // dh := h.Perc * math.Exp(dwt) * (1. - h.Fimp)
// 		// // if -avail > dh { // (Note: avail and dh cannot be >0.)
// 		// // 	avail += dh      // remaining available ep (cannot be >0.)
// 		// // 	rch -= dh        // (Note: dh cannot be >0.)
// 		// // 	aet = ep + avail // actual et
// 		// // } else {
// 		// // 	rch += avail // ep assumed unlimited from a saturated surface (Note: avail cannot be >0.)
// 		// // 	aet = ep     // completely satisfied over a high watertable
// 		// // }
// 		// // // option 4 limited evap 2
// 		// // dwt *= (1. - h.Fimp)
// 		// // if avail > dwt {
// 		// // 	rch += avail // ep assumed unlimited from a saturated surface (Note: avail cannot be >0.)
// 		// // 	aet = ep     // completely satisfied over a high watertable
// 		// // } else {
// 		// // 	avail -= dwt     // remaining available ep (cannot be >0.)
// 		// // 	rch += dwt       // (Note: dwt cannot be >0.)
// 		// // 	aet = ep + avail // actual et
// 		// // }
// 	} else {
// 		aet, ro, rch = h.Update(p, ep)
// 	}
// 	return
// }

// InfiltrateSurplus excess mobile water in infiltrated assuming a falling head through a unit length, returns added recharge
func (h *HRU) InfiltrateSurplus() float64 {
	d := -h.Sdet.Deficit()
	if d > 0 { // excess
		dh := d * (1. - math.Exp(-h.Perc))
		h.Sdet.Sto -= dh
		return dh
	}
	return 0.
}

// // // Update hru given a set of forcings
// // func (h *HRU) Update(p, ep float64) (aet, ro, rch float64) {
// // 	sri := h.Fimp * p // impervious runoff
// // 	ro = h.Sma.Overflow(h.Sdet.Overflow(p-sri)) + sri
// // 	rch = h.Sma.Overflow(-h.Perc) + h.Perc
// // 	avail := h.Sdet.Overflow(-ep) // remaining available ep
// // 	avail = h.Sma.Overflow(avail*h.Fprv) + avail*h.Fimp
// // 	aet = ep + avail
// // 	// h.updateStatus()
// // 	return
// // }

// // // UpdateWT hru given a set of forcings and the presence of a watertable
// // func (h *HRU) UpdateWT(p, ep, zwt float64) (aet, ro, rch float64) {
// // 	if zwt < 0. { // upward gradient
// // 		x := h.Sma.Skim() // excess water (note: the srf always overflows to sma)
// // 		if x < 0. {
// // 			rch = x               // groundwater discharge (negative recharge)
// // 			h.Sma.Sto = h.Sma.Cap // fill drainable porosity
// // 			x = 0.
// // 		}
// // 		sri := h.Fimp * p                    // impervious runoff
// // 		ro = h.Sdet.Overflow(p-sri) + sri + x // fulfill surface storage
// // 		aet = h.Fprv * ep                    // completely satisfied over a high watertable
// // 		rch -= aet                           // ep assumed equal to gw flux

// // 		// rch = -h.Sma.Deficit()           // groundwater discharge (negative recharge)
// // 		// h.Sma.Sto = h.Sma.Cap            // fill drainable porosity
// // 		// sri := h.Fimp * p                // impervious runoff
// // 		// ro = h.Sdet.Overflow(p-sri) + sri // fulfill surface storage
// // 		// rch += h.Sdet.Overflow(-ep)       // remaining available ep assumed to be taken from high watertable
// // 		// aet = ep
// // 	} else {
// // 		aet, ro, rch = h.Update(p, ep)
// // 	}
// // 	return
// // }

// // UpdateP adds precip to the hru and returns runoff
// func (h *HRU) UpdateP(p float64) float64 {
// 	sri := h.Fimp * p // impervious runoff
// 	return h.Sma.Overflow(h.Sdet.Overflow(p-sri)) + sri
// }

// // UpdateEp removes evaporation from hru
// func (h *HRU) UpdateEp(ep float64) float64 {
// 	avail := h.Sdet.Overflow(-ep) // remaining available ep
// 	avail = h.Sma.Overflow(avail*h.Fprv) + avail*h.Fimp
// 	return ep + avail
// }

// // UpdatePerc updates hru given no forcings (percolation only)
// func (h *HRU) UpdatePerc() float64 {
// 	return h.Sma.Overflow(-h.Perc) + h.Perc // amount recharged
// }

// // UpdatePercWT updates hru over a high water table
// // returns net groundwater exchange (negative: recharge, positive: discharge after filling soil zone)
// // zwt: depth to watertable [m]
// func (h *HRU) UpdatePercWT(zwt float64) float64 {
// 	if zwt < 0. { // upward gradient
// 		return h.Sma.Overflow(h.Perc * -zwt) // groundwater seepage, assumes unit gradient, unit thickness
// 	}
// 	return -(h.Sma.Overflow(-h.Perc) + h.Perc) // returns the amount recharged
// }

// // UpdateIN updates hru given a set of input forcings only
// func (h *HRU) UpdateIN(p float64) (ro, rch float64) {
// 	sri := h.Fimp * p // impervious runoff
// 	ro = h.Sma.Overflow(h.Sdet.Overflow(p-sri)) + sri
// 	rch = h.Sma.Overflow(-h.Perc) + h.Perc
// 	// h.updateStatus()
// 	return
// }

// // UpdateStorage simply adds infiltration to storage (soil first, excess to surface depressions)
// func (h *HRU) UpdateStorage(f float64) float64 {
// 	return h.Sdet.Overflow(h.Sma.Overflow(f))
// }

// // // AddToStorage simply adds infiltration to storage (soil first, excess to surface depressions), but keeps water onsite
// // func (h *HRU) AddToStorage(f float64) {
// // 	// h.Sdet.Sto += h.Sma.Overflow(f)
// // 	// h.Sma.Sto += h.Sdet.Overflow(f)
// // 	h.Sdet.Sto += f
// // }

// // // Storage2 returns total current storages (individually)
// // func (h *HRU) Storage2() (float64, float64) {
// // 	return h.Sma.Sto, h.Sdet.Sto
// // }

// // Deficit returns current storage deficit
// func (h *HRU) Deficit() float64 {
// 	return h.Sma.Deficit() + h.Sdet.Deficit()
// }

// // Infiltrability returns the amount of potential infiltration
// func (h *HRU) Infiltrability() float64 {
// 	return (h.Sma.Deficit() + h.Sdet.Deficit()) * h.Fprv
// }

// // func (h *HRU) updateStatus() {
// // 	if h.Sma.Sto < h.Sma.Cap {
// // 		h.stat |= deficitPoreWater
// // 		if h.Sma.Sto > 0. {
// // 			h.stat |= availPoreWater
// // 		} else {
// // 			h.stat &^= availPoreWater
// // 		}
// // 	} else {
// // 		h.stat &^= deficitPoreWater
// // 		if h.Sma.Sto > h.Sma.Cap {
// // 			h.stat |= waterOnSurface
// // 		} else {
// // 			h.stat &^= waterOnSurface
// // 		}
// // 	}
// // }
