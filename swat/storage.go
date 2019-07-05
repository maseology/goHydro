package swat

import (
	"log"
	"math"
)

// Storage returns the current moisture state [mm]
func (bsn *SubBasin) Storage() float64 {
	s := bsn.aq + bsn.psto + bsn.qstr + bsn.chn.vstr/bsn.Ca/1000.
	if math.IsNaN(s) {
		log.Fatalf("ERROR: SubBasin.Storage() is NaN")
	}
	for _, m := range bsn.hru {
		s += m.storage() * m.f
	}
	return s
}

// StorageAll returns the moisture states of all components [mm]
func (bsn *SubBasin) StorageAll() (aq, psto, qstr, vstr, sz float64) {
	aq, psto, qstr, vstr, sz = bsn.aq, bsn.psto, bsn.qstr, bsn.chn.vstr/bsn.Ca/1000., 0.
	for _, m := range bsn.hru {
		sz += m.storage() * m.f
	}
	return
}

func (m *HRU) storage() float64 {
	s := 0.
	for i := 0; i < nsl; i++ {
		s += m.sz[i].sw
	}
	if math.IsNaN(s) {
		log.Fatalf("ERROR: HRU.storage() is NaN")
	}
	return s
}

func (m *HRU) drainableStorage() float64 {
	swprfl := 0. // soil water content of the entire profile excluding the water held in the profile at wilting point [mm] (pg.104)
	for _, ly := range m.sz {
		if ly.sw > ly.wp {
			swprfl += ly.sw - ly.wp
		}
	}
	return swprfl
}
